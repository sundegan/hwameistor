package volume

import (
	"context"
	"fmt"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/node/disk"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/types"
	"sort"
	"strconv"

	"github.com/container-storage-interface/spec/lib/go/csi"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/builder/localdiskvolume"
	volumectr "github.com/hwameistor/hwameistor/pkg/local-disk-manager/handler/localdiskvolume"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils/kubernetes"
)

type DiskType = string

// consts
const (
	VolumeParameterDiskTypeKey     = "diskType"
	VolumeParameterMinCapacityKey  = "minCap"
	VolumeParameterPVCNameKey      = "csi.storage.k8s.io/pvc/name"
	VolumeParameterPVCNameSpaceKey = "csi.storage.k8s.io/pvc/namespace"
	VolumeSelectedNodeKey          = "volume.kubernetes.io/selected-node"
)

// localDiskVolumeManager manage the allocation, deletion and query of local disk data volumes.
// Internally, the reasonable allocation of data volumes will be realized by tuning the LocalDiskNode resources
type localDiskVolumeManager struct {
	// SupportVolumeCapacities
	SupportVolumeCapacities []*csi.VolumeCapability

	// dm manager all disks in cluster
	dm disk.Manager

	// GetClient for query LocalDiskVolume resources from k8s
	GetClient func() (*localdiskvolume.Kubeclient, error)

	// volume
	// The handler cannot be placed here directly as an object because thread safety cannot be guaranteed
	GetVolumeHandler func() (*volumectr.DiskVolumeHandler, error)
}

// VolumeRequest
type VolumeRequest struct {
	// RequireCapacity
	RequireCapacity int64 `json:"capacity"`

	// VolumeContext
	VolumeContext map[string]string `json:"volumeContext"`

	// DiskType represents which disk type is this volume provisioned from
	DiskType DiskType `json:"diskType"`

	// DevPath
	DevPath string `json:"devPath"`

	// PVCName
	PVCName string `json:"pvcName"`

	// PVCNameSpace
	PVCNameSpace string `json:"pvcNameSpace"`

	// OwnerNodeName represents where this disk volume located
	OwnerNodeName string `json:"ownerNodeName"`

	// VolumeCap
	VolumeCap *csi.VolumeCapability

	// VolumeContentSource
	// this field may be needed for volume clone from another disk volume
	// for now, we don't support this
	VolumeContentSource *csi.VolumeContentSource `json:"volumeContentSource"`
}

func NewVolumeRequest() *VolumeRequest {
	return &VolumeRequest{}
}

func (r *VolumeRequest) SetRequireCapacity(cap int64) {
	r.RequireCapacity = cap
}

func (r *VolumeRequest) SetPVCName(pvc string) {
	r.PVCName = pvc
}

func (r *VolumeRequest) SetPVCNameSpace(ns string) {
	r.PVCNameSpace = ns
}

func (r *VolumeRequest) SetNodeName(nodeName string) {
	r.OwnerNodeName = nodeName
}

func (r *VolumeRequest) SetDiskType(diskType string) {
	r.DiskType = diskType
}

func (r *VolumeRequest) Valid() error {
	if r.DiskType == "" {
		return fmt.Errorf("DevType is empty")
	}
	if r.PVCName == "" {
		return fmt.Errorf("PVCName is empty")
	}
	if r.OwnerNodeName == "" {
		return fmt.Errorf("SelectedNode is empty")
	}
	return nil
}

func New() Manager {
	vm := &localDiskVolumeManager{}
	vm.initVolumeCapacities()
	vm.initKubernetesClient()
	vm.initLocalDiskManager()
	vm.initLocalVolumeHandler()

	return vm
}

func (vm *localDiskVolumeManager) CreateVolume(name string, parameters interface{}) (*types.Volume, error) {
	volumeRequest, err := vm.ParseVolumeRequest(parameters)
	if err != nil {
		log.WithError(err).Error("Failed to ParseVolumeRequest")
		return nil, err
	}
	logCtx := log.Fields{
		"volume":           name,
		"node":             volumeRequest.OwnerNodeName,
		"pvcNamespaceName": volumeRequest.PVCNameSpace + "/" + volumeRequest.PVCName}

	// select suitable disk for the volume
	selectDisk, err := vm.findSuitableDisk(volumeRequest)
	if err != nil {
		log.WithFields(logCtx).WithError(err).Error("Failed to find suitable disk")
		return nil, err
	}
	if selectDisk == nil {
		err = fmt.Errorf("there is no suitable disk")
		return nil, err
	}
	log.WithFields(logCtx).Debugf("Succeed to select disk %s success", selectDisk.Name)

	// create LocalDiskVolume if not exist
	volume, err := localdiskvolume.NewBuilder().WithName(name).
		SetupDiskType(volumeRequest.DiskType).
		SetupDisk(selectDisk.DevPath).
		SetupLocalDiskName(selectDisk.Name).
		SetupAllocateCap(selectDisk.Capacity).
		SetupRequiredCapacityBytes(volumeRequest.RequireCapacity).
		SetupPVCNameSpaceName(volumeRequest.PVCNameSpace + "/" + volumeRequest.PVCName).
		SetupAccessibility(v1alpha1.AccessibilityTopology{Nodes: []string{volumeRequest.OwnerNodeName}}).
		SetupStatus(v1alpha1.VolumeStateCreated).Build()
	if err != nil {
		log.WithError(err).Error("Failed to build volume object")
		return nil, err
	}

	v, err := vm.createVolume(volume)
	if err != nil {
		log.WithError(err).Error("Failed to create volume")
		return nil, err
	}

	return &types.Volume{
		Name:     v.Name,
		Exist:    true,
		Capacity: v.Status.AllocatedCapacityBytes,
		Ready:    v.Status.State == v1alpha1.VolumeStateReady}, nil
}

func (vm *localDiskVolumeManager) UpdateVolume(name string, parameters interface{}) (*types.Volume, error) {
	r, err := vm.ParseVolumeRequest(parameters)
	if err != nil {
		log.WithError(err).Error("Failed to ParseVolumeRequest")
		return nil, err
	}

	volume, err := vm.getVolume(name)
	if err != nil {
		return nil, err
	}

	if volume.Status.AllocatedCapacityBytes < r.RequireCapacity {
		return nil, fmt.Errorf("RequireCapacity in VolumeRequest is modified "+
			"but is bigger than allocted disk %s/%s (the disk capacity %d)",
			volume.Spec.Accessibility.Nodes, volume.Status.DevPath, volume.Status.AllocatedCapacityBytes)
	}

	newVolume, err := localdiskvolume.NewBuilderFrom(volume).
		SetupAccessibility(v1alpha1.AccessibilityTopology{Nodes: []string{r.OwnerNodeName}}).
		SetupRequiredCapacityBytes(r.RequireCapacity).
		SetupDiskType(r.DiskType).
		SetupPVCNameSpaceName(r.PVCNameSpace + "/" + r.PVCName).Build()
	if err != nil {
		return nil, err
	}

	v, err := vm.updateVolume(newVolume)
	if err != nil {
		return nil, err
	}

	return &types.Volume{
		Name:     v.Name,
		Exist:    true,
		Capacity: v.Status.AllocatedCapacityBytes,
		Ready:    v.Status.State == v1alpha1.VolumeStateReady}, nil
}

func (vm *localDiskVolumeManager) newHandlerForVolume(name string) (*volumectr.DiskVolumeHandler, error) {
	vh, err := vm.GetVolumeHandler()
	if err != nil {
		return nil, err
	}
	volume, err := vm.getVolume(name)
	if err != nil {
		return nil, err
	}
	vh.For(volume)
	return vh, nil
}

func (vm *localDiskVolumeManager) NodePublishVolume(ctx context.Context, volumeReq interface{}) error {
	r, ok := volumeReq.(*csi.NodePublishVolumeRequest)
	if !ok {
		return fmt.Errorf("NodePublishRequest is not valid")
	}

	volume, err := vm.newHandlerForVolume(r.GetVolumeId())
	if err != nil {
		return err
	}

	// update mountPoint to LocalVolume
	exist := volume.ExistMountPoint(r.GetTargetPath())
	if !exist {
		volume.AppendMountPoint(r.GetTargetPath(), r.GetVolumeCapability())
		volume.SetupVolumeStatus(v1alpha1.VolumeStateNotReady)

		if err = volume.UpdateLocalDiskVolume(); err != nil {
			return err
		}
	}

	return volume.WaitVolume(ctx, v1alpha1.VolumeStateReady)
}

func (vm *localDiskVolumeManager) NodeUnpublishVolume(ctx context.Context,
	name, targetPath string) error {
	volume, err := vm.newHandlerForVolume(name)
	if err != nil {
		if errors.IsNotFound(err) {
			log.WithFields(log.Fields{"Volume": name, "TargetPath": targetPath}).Errorf(
				"LocalDiskVolume has been deleted for some unknown reason, "+
					"you may need to umount it manually, "+
					"cmd: umount %s", targetPath)
			return nil
		}
		return err
	}

	volume.UpdateMountPointPhase(targetPath, v1alpha1.MountPointToBeUnMount)
	volume.SetupVolumeStatus(v1alpha1.VolumeStateToBeUnmount)
	if err := volume.UpdateLocalDiskVolume(); err != nil {
		return err
	}

	return volume.WaitVolumeUnmounted(ctx, targetPath)
}

func (vm *localDiskVolumeManager) DeleteVolume(ctx context.Context, name string) error {
	volume, err := vm.newHandlerForVolume(name)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Infof("Volume %s is deleted already", name)
			return nil
		}
		return err
	}

	// 1. wait all volume unmounted
	if len(volume.GetMountPoints()) > 0 {
		err = fmt.Errorf("volume %s has %d mountpoint remained, can't delete volume now",
			name, len(volume.GetMountPoints()))
		log.WithError(err).Error("Failed to delete volume")
		return err
	}

	// fixme: The ToBeDeleted status seems a little redundant, and nothing is actually done.
	//  If it is to check the status of the data volume associated with the disk, it seems that the mount point is enough
	// 1.1 update volume state to ToBeDeleted
	// this step is ensure all mountpoints are safely umount
	volume.SetupVolumeStatus(v1alpha1.VolumeStateToBeDeleted)
	if err = volume.UpdateLocalDiskVolume(); err != nil {
		log.WithError(err).Error("Failed to delete volume")
		return err
	}
	if err = volume.WaitVolume(ctx, v1alpha1.VolumeStateDeleted); err != nil {
		log.WithError(err).Error("Failed to delete volume")
		return err
	}

	// 2. once volume is safely deleted, disk can be released
	// todo: release disk

	// 3. remove finalizer, volume will be deleted totally
	_ = volume.RemoveFinalizers()
	return volume.UpdateLocalDiskVolume()
}

func (vm *localDiskVolumeManager) GetVolumeInfo(name string) (*types.Volume, error) {
	volume := &types.Volume{}
	exist, err := vm.VolumeIsExist(name)
	if err != nil {
		return nil, err
	}
	volume.Exist = exist

	if !volume.Exist {
		return volume, nil
	}

	v, err := vm.getVolume(name)
	if err != nil {
		return nil, err
	}
	volume.Name = v.GetName()
	volume.Capacity = v.Status.AllocatedCapacityBytes
	if len(v.Spec.Accessibility.Nodes) > 0 {
		volume.AttachNode = v.Spec.Accessibility.Nodes[0]
	}

	return volume, nil
}

func (vm *localDiskVolumeManager) VolumeIsReady(name string) (bool, error) {
	vol, err := vm.getVolume(name)
	if err != nil {
		log.WithError(err).Error("Failed to get disk volume")
		return false, err
	}

	return vol.Status.State == v1alpha1.VolumeStateReady, nil
}

func (vm *localDiskVolumeManager) VolumeIsExist(name string) (bool, error) {
	vol, err := vm.getVolume(name)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return vol.Name == name, nil
}

func (vm *localDiskVolumeManager) GetVolumeCapacities() interface{} {
	return vm.SupportVolumeCapacities
}

func (vm *localDiskVolumeManager) initKubernetesClient() {
	vm.GetClient = localdiskvolume.NewKubeclient
}

func (vm *localDiskVolumeManager) initLocalDiskManager() {
	vm.dm = disk.New()
}

func (vm *localDiskVolumeManager) initLocalVolumeHandler() {
	client, err := kubernetes.NewClient()
	if err != nil {
		log.WithError(err).Error("Failed to new kubernetes client")
		return
	}

	recorder, err := kubernetes.NewRecorderFor("localdisk-volumemanager")
	if err != nil {
		log.WithError(err).Error("Failed to new kubernetes recorder")
		return
	}

	vm.GetVolumeHandler = func() (*volumectr.DiskVolumeHandler, error) {
		if client == nil || recorder == nil {
			return nil, fmt.Errorf("failed to get DiskVolumeHandler, object is empty")
		}
		return volumectr.NewLocalDiskVolumeHandler(client, recorder), nil
	}
}

func (vm *localDiskVolumeManager) initVolumeCapacities() {
	vm.SupportVolumeCapacities = []*csi.VolumeCapability{
		{ // Tell CO we can provision readWriteOnce raw block volumes.
			AccessType: &csi.VolumeCapability_Block{
				Block: &csi.VolumeCapability_BlockVolume{},
			},
			AccessMode: &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			},
		},
		{ // Tell CO we can provision readWriteOnce filesystem volumes.
			AccessType: &csi.VolumeCapability_Mount{
				Mount: &csi.VolumeCapability_MountVolume{},
			},
			AccessMode: &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			},
		},
	}
	for _, c := range vm.SupportVolumeCapacities {
		log.WithField("capability", c).Info("Enabling volume capability")
	}
}

// ParseVolumeRequest ParseParams
func (vm *localDiskVolumeManager) ParseVolumeRequest(parameters interface{}) (*VolumeRequest, error) {
	r, ok := parameters.(*csi.CreateVolumeRequest)
	if !ok {
		return nil, fmt.Errorf("volume request type error, not the CreateVolumeRequest")
	}

	logBase := log.WithFields(utils.StructToMap(r, "json"))
	volumeRequest := &VolumeRequest{
		VolumeContext:       r.GetParameters(),
		VolumeContentSource: r.VolumeContentSource,
	}

	// check volume Capabilities
	if _, err := vm.isSupportVolumeCapabilities(r.GetVolumeCapabilities()); err != nil {
		logBase.WithError(err).Error("Failed to check VolumeCapabilities")
		return nil, err
	}

	volumeRequest.SetDiskType(r.GetParameters()[VolumeParameterDiskTypeKey])
	volumeRequest.SetPVCName(r.GetParameters()[VolumeParameterPVCNameKey])
	volumeRequest.SetPVCNameSpace(r.GetParameters()[VolumeParameterPVCNameSpaceKey])
	if r.AccessibilityRequirements != nil &&
		len(r.AccessibilityRequirements.Requisite) == 1 {
		if nodeName, ok := r.AccessibilityRequirements.Requisite[0].Segments[TopologyNodeKey]; ok {
			volumeRequest.SetNodeName(nodeName)
		}
	}
	requireBytes, err := vm.quireBytes(r)
	if err != nil {
		log.WithError(err).Error("Failed to parse RequireBytes")
		return nil, err
	}

	volumeRequest.SetRequireCapacity(requireBytes)
	return volumeRequest, volumeRequest.Valid()
}

// isSupportVolumeCapability
func (vm *localDiskVolumeManager) isSupportVolumeCapabilities(caps []*csi.VolumeCapability) (bool, error) {
	supportCaps, ok := vm.GetVolumeCapacities().([]*csi.VolumeCapability)
	if !ok {
		log.WithFields(utils.StructToMap(vm.GetVolumeCapacities(), "json")).Error("Failed to get VolumeCapacities")
		return false, fmt.Errorf("failed to get VolumeCapacities")
	}

	// check AccessMode
	for _, needCap := range caps {
		support := false
		for _, supportCap := range supportCaps {
			if supportCap.GetAccessMode().GetMode() == needCap.GetAccessMode().GetMode() {
				support = true
				break
			}
		}

		if !support {
			return false, fmt.Errorf("don't support VolumeCapability %s", needCap.String())
		}
	}

	return true, nil
}

func (vm *localDiskVolumeManager) getVolume(name string) (*v1alpha1.LocalDiskVolume, error) {
	client, err := vm.GetClient()
	if err != nil {
		return nil, err
	}

	return client.Get(name)
}

func (vm *localDiskVolumeManager) createVolume(volume *v1alpha1.LocalDiskVolume) (*v1alpha1.LocalDiskVolume, error) {
	client, err := vm.GetClient()
	if err != nil {
		log.WithError(err).Error("Failed to create kubernetes client")
		return nil, err
	}

	return client.Create(volume)
}

func (vm *localDiskVolumeManager) updateVolume(volume *v1alpha1.LocalDiskVolume) (*v1alpha1.LocalDiskVolume, error) {
	client, err := vm.GetClient()
	if err != nil {
		return nil, err
	}

	return client.Update(volume)
}

func (vm *localDiskVolumeManager) quireBytes(csiRequest *csi.CreateVolumeRequest) (int64, error) {
	pvcRequireBytes := int64(0)
	if csiRequest.GetCapacityRange() != nil {
		pvcRequireBytes = csiRequest.GetCapacityRange().GetRequiredBytes()
	}

	scRequireBytes := int64(0)
	var err error
	var base, bitSize = 10, 64
	if minCap, ok := csiRequest.GetParameters()[VolumeParameterMinCapacityKey]; ok {
		if scRequireBytes, err = strconv.ParseInt(minCap, base, bitSize); err != nil {
			log.WithError(err).Error("Parse min cap from StorageClass fail")
		}
	}

	if pvcRequireBytes < scRequireBytes {
		return scRequireBytes, nil
	}

	if pvcRequireBytes <= 0 {
		return pvcRequireBytes, fmt.Errorf("RequireBytes is less than 0 Bytes")
	}

	return pvcRequireBytes, nil
}

// findSuitableDisk according volume request(contains attach-node, request storage capacity)
func (vm *localDiskVolumeManager) findSuitableDisk(vq *VolumeRequest) (*types.Disk, error) {
	disks, err := vm.dm.GetNodeDisks(vq.OwnerNodeName)
	if err != nil {
		return nil, err
	}
	var suitableDisk *types.Disk
	var sortDisks = utils.ByDiskSize(disks)
	sort.Sort(sortDisks)
	for _, d := range sortDisks {
		// todo: filter disk already reserve by volume according to localdiskvolume
		if d.Status != types.DiskStatusAvailable || d.DiskType != vq.DiskType || d.Capacity < vq.RequireCapacity {
			continue
		}
		suitableDisk = &d
	}
	return suitableDisk, nil
}
