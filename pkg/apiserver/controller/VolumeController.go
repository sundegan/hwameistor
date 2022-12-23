package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	hwameistorapi "github.com/hwameistor/hwameistor/pkg/apiserver/api"
	"github.com/hwameistor/hwameistor/pkg/apiserver/manager"
)

type IVolumeController interface {
	//RestController
	VolumeGet(ctx *gin.Context)
	VolumeList(ctx *gin.Context)
	VolumeReplicasGet(ctx *gin.Context)
	GetVolumeMigrateOperation(ctx *gin.Context)
	VolumeMigrateOperation(ctx *gin.Context)
	GetVolumeConvertOperation(ctx *gin.Context)
	VolumeConvertOperation(ctx *gin.Context)
	GetVolumeExpandOperation(ctx *gin.Context)
	VolumeExpandOperation(ctx *gin.Context)
}

// VolumeController
type VolumeController struct {
	m *manager.ServerManager
}

func NewVolumeController(m *manager.ServerManager) IVolumeController {
	fmt.Println("NewVolumeController start")

	return &VolumeController{m}
}

// VolumeGet godoc
// @Summary     摘要 获取指定数据卷基本信息
// @Description get Volume angel1
// @Tags        Volume
// @Param       volumeName path string true "volumeName"
// @Accept      application/json
// @Produce     application/json
// @Success     200 {object} api.Volume
// @Router      /cluster/volumes/{volumeName} [get]
func (v *VolumeController) VolumeGet(ctx *gin.Context) {
	// 获取path中的name
	volumeName := ctx.Param("volumeName")

	if volumeName == "" {
		ctx.JSON(http.StatusNonAuthoritativeInfo, nil)
		return
	}
	lv, err := v.m.VolumeController().GetLocalVolume(volumeName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, lv)
}

// VolumeList godoc
// @Summary 摘要 获取数据卷列表信息
// @Description list Volume
// @Tags        Volume
// @Param       page query int32 true "page"
// @Param       pageSize query int32 true "pageSize"
// @Param       volumeName query string false "volumeName"
// @Param       state query string false "state"
// @Param       nameSpace query string false "nameSpace"
// @Param       fuzzy query bool false "fuzzy"
// @Param       sort query bool false "sort"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.VolumeList   "成功"
// @Router      /cluster/volumes [get]
func (v *VolumeController) VolumeList(ctx *gin.Context) {

	// 获取path中的page
	page := ctx.Query("page")
	// 获取path中的pageSize
	pageSize := ctx.Query("pageSize")
	// 获取path中的volumeName
	volumeName := ctx.Query("volumeName")
	fmt.Println("VolumeList volumeName = %v", volumeName)
	// 获取path中的state
	state := ctx.Query("state")
	// 获取path中的namespace
	namespace := ctx.Query("namespace")

	p, _ := strconv.ParseInt(page, 10, 32)
	ps, _ := strconv.ParseInt(pageSize, 10, 32)

	var queryPage hwameistorapi.QueryPage
	queryPage.Page = int32(p)
	queryPage.PageSize = int32(ps)
	queryPage.VolumeName = volumeName
	queryPage.VolumeState = hwameistorapi.VolumeStatefuzzyConvert(state)
	queryPage.NameSpace = namespace

	lvs, err := v.m.VolumeController().ListLocalVolume(queryPage)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, lvs)
}

// VolumeReplicasGet godoc
// @Summary 摘要 获取指定数据卷的副本列表信息
// @Description list volumes
// @Tags        Volume
// @Param       volumeName path string true "volumeName"
// @Param       volumeReplicaName query string false "volumeReplicaName"
// @Param       state query string false "state"
// @Param       fuzzy query bool false "fuzzy"
// @Param       sort query bool false "sort"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.VolumeReplicaList  "成功"
// @Router      /cluster/volumes/{volumeName}/replicas [get]
func (v *VolumeController) VolumeReplicasGet(ctx *gin.Context) {
	// 获取path中的name
	volumeName := ctx.Param("volumeName")

	// 获取path中的volumeReplicaName
	volumeReplicaName := ctx.Query("volumeReplicaName")
	// 获取path中的state
	state := ctx.Query("state")

	if volumeName == "" {
		ctx.JSON(http.StatusNonAuthoritativeInfo, nil)
		return
	}

	var queryPage hwameistorapi.QueryPage
	queryPage.VolumeReplicaName = volumeReplicaName
	queryPage.VolumeState = hwameistorapi.VolumeStatefuzzyConvert(state)
	queryPage.VolumeName = volumeName

	lvs, err := v.m.VolumeController().GetVolumeReplicas(queryPage)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, lvs)
}

// VolumeMigrateOperation godoc
// @Summary 摘要 指定数据卷迁移操作
// @Description post VolumeMigrateOperation body i.g. body { SrcNode string ,SelectedNode string}
// @Tags        Volume
// @Param       volumeName path string true "volumeName"
// @Param       body body api.VolumeMigrateReqBody true "reqBody"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.VolumeMigrateRspBody      "成功"
// @Failure     500 {object}  api.RspFailBody "失败"
// @Router      /cluster/volumes/{volumeName}/migrate [post]
func (v *VolumeController) VolumeMigrateOperation(ctx *gin.Context) {

	// 获取path中的name
	name := ctx.Param("volumeName")
	fmt.Println("VolumeMigrateOperation name = %v", name)

	if name == "" {
		ctx.JSON(http.StatusNonAuthoritativeInfo, nil)
		return
	}

	var vmrb hwameistorapi.VolumeMigrateReqBody
	err := ctx.ShouldBind(&vmrb)
	if err != nil {
		fmt.Errorf("Unmarshal err = %v", err)
		ctx.JSON(http.StatusNonAuthoritativeInfo, nil)
		return
	}

	sourceNodeName := vmrb.SrcNode
	targetNodeName := vmrb.SelectedNode
	abort := vmrb.Abort

	fmt.Println("VolumeMigrateOperation sourceNodeName = %v, targetNodeName = %v", sourceNodeName, targetNodeName)

	volumeMigrate, err := v.m.VolumeController().CreateVolumeMigrate(name, sourceNodeName, targetNodeName, abort)
	if err != nil {
		var failRsp hwameistorapi.RspFailBody
		failRsp.ErrCode = 500
		failRsp.Desc = "VolumeMigrateOperation Failed: " + err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}
	ctx.JSON(http.StatusOK, volumeMigrate)
}

// VolumeConvertOperation godoc
// @Summary 摘要 指定数据卷转换操作
// @Description post VolumeConvertOperation
// @Tags        Volume
// @Param       volumeName path string true "volumeName"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.VolumeConvertRspBody      "成功"
// @Failure     500 {object}  api.RspFailBody "失败"
// @Router      /cluster/volumes/{volumeName}/convert [post]
func (v *VolumeController) VolumeConvertOperation(ctx *gin.Context) {

	// 获取path中的name
	volumeName := ctx.Param("volumeName")

	//var vcrb hwameistorapi.VolumeConvertReqBody
	//err := ctx.ShouldBind(&vcrb)
	//if err != nil {
	//	fmt.Errorf("Unmarshal err = %v", err)
	//	ctx.JSON(http.StatusNonAuthoritativeInfo, nil)
	//	return
	//}
	//volumeName := vcrb.VolumeName

	fmt.Println("VolumeConvertOperation volumeName = %v", volumeName)
	if volumeName == "" {
		ctx.JSON(http.StatusNonAuthoritativeInfo, nil)
		return
	}

	volumeConvert, err := v.m.VolumeController().CreateVolumeConvert(volumeName)
	if err != nil {
		var failRsp hwameistorapi.RspFailBody
		failRsp.ErrCode = 500
		failRsp.Desc = "VolumeConvertOperation Failed: " + err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, volumeConvert)
}

// GetVolumeMigrateOperation godoc
// @Summary 摘要 获取指定数据卷迁移操作
// @Description get GetVolumeMigrateOperation
// @Tags        Volume
// @Param       volumeName path string true "volumeName"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.VolumeMigrateOperation      "成功"
// @Failure     500 {object}  api.RspFailBody "失败"
// @Router      /cluster/volumes/{volumeName}/migrate [get]
func (v *VolumeController) GetVolumeMigrateOperation(ctx *gin.Context) {

	// 获取path中的name
	volumeName := ctx.Param("volumeName")

	fmt.Println("VolumeConvertOperation volumeName = %v", volumeName)
	if volumeName == "" {
		ctx.JSON(http.StatusNonAuthoritativeInfo, nil)
		return
	}

	volumeConvert, err := v.m.VolumeController().GetVolumeMigrate(volumeName)
	if err != nil {
		var failRsp hwameistorapi.RspFailBody
		failRsp.ErrCode = 500
		failRsp.Desc = "VolumeConvertOperation Failed: " + err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, volumeConvert)
}

// GetVolumeConvertOperation godoc
// @Summary 摘要 获取指定数据卷转换操作
// @Description get GetVolumeConvertOperation
// @Tags        Volume
// @Param       volumeName path string true "volumeName"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.VolumeConvertOperation      "成功"
// @Failure     500 {object}  api.RspFailBody "失败"
// @Router      /cluster/volumes/{volumeName}/convert [get]
func (v *VolumeController) GetVolumeConvertOperation(ctx *gin.Context) {

	// 获取path中的name
	volumeName := ctx.Param("volumeName")

	fmt.Println("VolumeConvertOperation volumeName = %v", volumeName)
	if volumeName == "" {
		ctx.JSON(http.StatusNonAuthoritativeInfo, nil)
		return
	}

	volumeConvert, err := v.m.VolumeController().GetVolumeConvert(volumeName)
	if err != nil {
		var failRsp hwameistorapi.RspFailBody
		failRsp.ErrCode = 500
		failRsp.Desc = "VolumeConvertOperation Failed: " + err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, volumeConvert)
}

// GetVolumeExpandOperation godoc
// @Summary 摘要 获取指定数据卷扩容操作
// @Description get GetVolumeExpandOperation
// @Tags        Volume
// @Param       volumeName path string true "volumeName"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.VolumeExpandOperation      "成功"
// @Failure     500 {object}  api.RspFailBody "失败"
// @Router      /cluster/volumes/{volumeName}/expand [get]
func (v *VolumeController) GetVolumeExpandOperation(ctx *gin.Context) {

	// 获取path中的name
	volumeName := ctx.Param("volumeName")

	fmt.Println("VolumeConvertOperation volumeName = %v", volumeName)
	if volumeName == "" {
		ctx.JSON(http.StatusNonAuthoritativeInfo, nil)
		return
	}

	volumeConvert, err := v.m.VolumeController().GetVolumeConvert(volumeName)
	if err != nil {
		var failRsp hwameistorapi.RspFailBody
		failRsp.ErrCode = 500
		failRsp.Desc = "VolumeConvertOperation Failed: " + err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, volumeConvert)
}

// VolumeExpandOperation godoc
// @Summary 摘要 指定数据卷转换
// @Description post VolumeExpandOperation
// @Tags        Volume
// @Param       volumeName path string true "volumeName"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.VolumeConvertOperation      "成功"
// @Failure     500 {object}  api.RspFailBody "失败"
// @Router      /cluster/volumes/{volumeName}/expand [post]
func (v *VolumeController) VolumeExpandOperation(ctx *gin.Context) {

	// 获取path中的name
	volumeName := ctx.Param("volumeName")

	fmt.Println("VolumeConvertOperation volumeName = %v", volumeName)
	if volumeName == "" {
		ctx.JSON(http.StatusNonAuthoritativeInfo, nil)
		return
	}

	volumeConvert, err := v.m.VolumeController().GetVolumeConvert(volumeName)
	if err != nil {
		var failRsp hwameistorapi.RspFailBody
		failRsp.ErrCode = 500
		failRsp.Desc = "VolumeConvertOperation Failed: " + err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, volumeConvert)
}
