package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/middleware"
	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	apperrors "github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/response"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
	"github.com/luoxiaojun1992/miniprogram/internal/service"
)

// CourseController handles course requests.
type CourseController struct {
	svc            service.CourseService
	log            *logrus.Logger
	accessChecker  *accessChecker
	courseRepo     repository.CourseRepository
	courseUnitRepo repository.CourseUnitRepository
	attachRepo     repository.CourseAttachmentRepository
	unitAttachRepo repository.CourseUnitAttachmentRepository
}

// NewCourseController creates a new CourseController.
func NewCourseController(
	svc service.CourseService,
	log *logrus.Logger,
	deps ...interface{},
) *CourseController {
	var contentPermRepo repository.ContentPermissionRepository
	var roleRepo repository.RoleRepository
	var courseRepo repository.CourseRepository
	var courseUnitRepo repository.CourseUnitRepository
	var attachRepo repository.CourseAttachmentRepository
	var unitAttachRepo repository.CourseUnitAttachmentRepository
	for _, dep := range deps {
		switch v := dep.(type) {
		case repository.ContentPermissionRepository:
			contentPermRepo = v
		case repository.RoleRepository:
			roleRepo = v
		case repository.CourseRepository:
			courseRepo = v
		case repository.CourseUnitRepository:
			courseUnitRepo = v
		case repository.CourseAttachmentRepository:
			attachRepo = v
		case repository.CourseUnitAttachmentRepository:
			unitAttachRepo = v
		}
	}
	return &CourseController{
		svc:            svc,
		log:            log,
		accessChecker:  newAccessChecker(contentPermRepo, roleRepo),
		courseRepo:     courseRepo,
		courseUnitRepo: courseUnitRepo,
		attachRepo:     attachRepo,
		unitAttachRepo: unitAttachRepo,
	}
}

// List handles GET /courses.
func (c *CourseController) List(ctx *gin.Context) {
	var q dto.ListQuery
	if err := ctx.ShouldBindQuery(&q); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	var moduleID *uint
	if m := ctx.Query("module_id"); m != "" {
		v, _ := strconv.ParseUint(m, 10, 32)
		u := uint(v)
		moduleID = &u
	}
	var isFree *bool
	if f := ctx.Query("is_free"); f != "" {
		b := f == "true" || f == "1"
		isFree = &b
	}
	userID, _ := middleware.GetCurrentUserID(ctx)
	var uid *uint64
	if userID > 0 {
		uid = &userID
	}
	courses, total, err := c.svc.List(ctx, q.GetPage(), q.GetPageSize(), q.Keyword, moduleID, isFree, uid)
	if err != nil {
		ctx.Error(err)
		return
	}
	filtered := make([]*entity.Course, 0, len(courses))
	for _, item := range courses {
		allowed, accessErr := c.accessChecker.canAccess(ctx, 3, uint64(item.ModuleID), uid, nil)
		if accessErr != nil {
			ctx.Error(accessErr)
			return
		}
		if allowed {
			filtered = append(filtered, item)
		}
	}
	response.PaginatedSuccess(ctx, filtered, minInt64(total, int64(len(filtered))), q.GetPage(), q.GetPageSize())
}

// GetByID handles GET /courses/:id.
func (c *CourseController) GetByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的课程ID", err))
		return
	}
	userID, _ := middleware.GetCurrentUserID(ctx)
	var uid *uint64
	if userID > 0 {
		uid = &userID
	}
	course, svcErr := c.svc.GetByID(ctx, id, uid)
	if svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.Success(ctx, course)
}

// AdminList handles GET /admin/courses.
func (c *CourseController) AdminList(ctx *gin.Context) {
	var q dto.ListQuery
	if err := ctx.ShouldBindQuery(&q); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	var status *int8
	if s := ctx.Query("status"); s != "" {
		v, _ := strconv.ParseInt(s, 10, 8)
		t := int8(v)
		status = &t
	}
	courses, total, err := c.svc.AdminList(ctx, q.GetPage(), q.GetPageSize(), q.Keyword, status)
	if err != nil {
		ctx.Error(err)
		return
	}
	response.PaginatedSuccess(ctx, courses, total, q.GetPage(), q.GetPageSize())
}

// AdminGetByID handles GET /admin/courses/:id.
func (c *CourseController) AdminGetByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的课程ID", err))
		return
	}
	course, svcErr := c.svc.AdminGetByID(ctx, id)
	if svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.Success(ctx, course)
}

// AdminCreate handles POST /admin/courses.
func (c *CourseController) AdminCreate(ctx *gin.Context) {
	var req dto.CreateCourseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	if err := req.Validate(); err != nil {
		ctx.Error(apperrors.NewValidation("参数校验失败", err))
		return
	}
	authorID, _ := middleware.GetCurrentUserID(ctx)
	id, err := c.svc.Create(ctx, &req, authorID)
	if err != nil {
		ctx.Error(err)
		return
	}
	response.SuccessWithStatus(ctx, http.StatusCreated, gin.H{"id": id})
}

// AdminUpdate handles PUT /admin/courses/:id.
func (c *CourseController) AdminUpdate(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的课程ID", err))
		return
	}
	var req dto.UpdateCourseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	if err := req.Validate(); err != nil {
		ctx.Error(apperrors.NewValidation("参数校验失败", err))
		return
	}
	if svcErr := c.svc.Update(ctx, id, &req); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.Success(ctx, nil)
}

// AdminDelete handles DELETE /admin/courses/:id.
func (c *CourseController) AdminDelete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的课程ID", err))
		return
	}
	if svcErr := c.svc.Delete(ctx, id); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// AdminPublish handles POST /admin/courses/:id/publish.
func (c *CourseController) AdminPublish(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的课程ID", err))
		return
	}
	var req dto.PublishCourseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	if err := req.Validate(); err != nil {
		ctx.Error(apperrors.NewValidation("参数校验失败", err))
		return
	}
	if svcErr := c.svc.Publish(ctx, id, &req); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.Success(ctx, nil)
}

// AdminPin handles POST /admin/courses/:id/pin.
func (c *CourseController) AdminPin(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的课程ID", err))
		return
	}
	var req dto.PinCourseRequest
	if err = ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	if err = req.Validate(); err != nil {
		ctx.Error(apperrors.NewValidation("参数校验失败", err))
		return
	}
	if svcErr := c.svc.Pin(ctx, id, &req); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.Success(ctx, nil)
}

// AdminCopy handles POST /admin/courses/:id/copy.
func (c *CourseController) AdminCopy(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的课程ID", err))
		return
	}
	authorID, _ := middleware.GetCurrentUserID(ctx)
	newID, svcErr := c.svc.Copy(ctx, id, authorID)
	if svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.SuccessWithStatus(ctx, http.StatusCreated, gin.H{"id": newID})
}

// AdminGetUnits handles GET /admin/courses/:id/units.
func (c *CourseController) AdminGetUnits(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的课程ID", err))
		return
	}
	units, svcErr := c.svc.GetUnits(ctx, id)
	if svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.Success(ctx, units)
}

// AdminCreateUnit handles POST /admin/courses/:id/units.
func (c *CourseController) AdminCreateUnit(ctx *gin.Context) {
	courseID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的课程ID", err))
		return
	}
	var req dto.CreateCourseUnitRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	if err := req.Validate(); err != nil {
		ctx.Error(apperrors.NewValidation("参数校验失败", err))
		return
	}
	unitID, svcErr := c.svc.CreateUnit(ctx, courseID, &req)
	if svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.SuccessWithStatus(ctx, http.StatusCreated, gin.H{"id": unitID})
}

// AdminUpdateUnit handles PUT /admin/courses/:id/units/:unit_id.
func (c *CourseController) AdminUpdateUnit(ctx *gin.Context) {
	courseID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的课程ID", err))
		return
	}
	unitID, err := strconv.ParseUint(ctx.Param("unit_id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的单元ID", err))
		return
	}
	var req dto.CreateCourseUnitRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	if err := req.Validate(); err != nil {
		ctx.Error(apperrors.NewValidation("参数校验失败", err))
		return
	}
	if svcErr := c.svc.UpdateUnit(ctx, courseID, unitID, &req); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.Success(ctx, nil)
}

// AdminDeleteUnit handles DELETE /admin/courses/:id/units/:unit_id.
func (c *CourseController) AdminDeleteUnit(ctx *gin.Context) {
	courseID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的课程ID", err))
		return
	}
	unitID, err := strconv.ParseUint(ctx.Param("unit_id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的单元ID", err))
		return
	}
	if svcErr := c.svc.DeleteUnit(ctx, courseID, unitID); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// GetUnits handles GET /courses/:id/units.
func (c *CourseController) GetUnits(ctx *gin.Context) {
	courseID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的课程ID", err))
		return
	}
	if c.courseRepo == nil {
		units, svcErr := c.svc.GetUnits(ctx, courseID)
		if svcErr != nil {
			ctx.Error(svcErr)
			return
		}
		response.Success(ctx, units)
		return
	}
	course, err := c.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		ctx.Error(err)
		return
	}
	if course == nil {
		ctx.Error(apperrors.NewNotFound("课程不存在", nil))
		return
	}
	userID, _ := middleware.GetCurrentUserID(ctx)
	var uid *uint64
	if userID > 0 {
		uid = &userID
	}
	allowed, accessErr := c.accessChecker.canAccess(ctx, 2, courseID, uid, &course.AuthorID)
	if accessErr != nil {
		ctx.Error(accessErr)
		return
	}
	if !allowed {
		ctx.Error(apperrors.NewForbidden("无权限访问该课程单元", nil))
		return
	}
	units, svcErr := c.svc.GetUnits(ctx, courseID)
	if svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.Success(ctx, units)
}

// GetAttachments handles GET /courses/:id/attachments.
func (c *CourseController) GetAttachments(ctx *gin.Context) {
	courseID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的课程ID", err))
		return
	}
	if c.courseRepo == nil || c.attachRepo == nil {
		ctx.Error(apperrors.NewBadRequest("课程附件仓储未初始化", nil))
		return
	}
	course, err := c.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		ctx.Error(err)
		return
	}
	if course == nil {
		ctx.Error(apperrors.NewNotFound("课程不存在", nil))
		return
	}
	userID, _ := middleware.GetCurrentUserID(ctx)
	var uid *uint64
	if userID > 0 {
		uid = &userID
	}
	allowed, accessErr := c.accessChecker.canAccess(ctx, 2, courseID, uid, &course.AuthorID)
	if accessErr != nil {
		ctx.Error(accessErr)
		return
	}
	if !allowed {
		ctx.Error(apperrors.NewForbidden("无权限访问该课程附件", nil))
		return
	}
	fileIDs, repoErr := c.attachRepo.ListFileIDs(ctx, courseID)
	if repoErr != nil {
		ctx.Error(repoErr)
		return
	}
	response.Success(ctx, fileIDs)
}

// GetUnitAttachments handles GET /courses/:id/units/:unit_id/attachments.
func (c *CourseController) GetUnitAttachments(ctx *gin.Context) {
	courseID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的课程ID", err))
		return
	}
	unitID, err := strconv.ParseUint(ctx.Param("unit_id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的单元ID", err))
		return
	}
	if c.courseRepo == nil || c.courseUnitRepo == nil || c.unitAttachRepo == nil {
		ctx.Error(apperrors.NewBadRequest("课程单元附件仓储未初始化", nil))
		return
	}
	unit, repoErr := c.courseUnitRepo.GetByID(ctx, unitID)
	if repoErr != nil {
		ctx.Error(repoErr)
		return
	}
	if unit == nil || unit.CourseID != courseID {
		ctx.Error(apperrors.NewNotFound("课程单元不存在", nil))
		return
	}
	course, repoErr := c.courseRepo.GetByID(ctx, courseID)
	if repoErr != nil {
		ctx.Error(repoErr)
		return
	}
	if course == nil {
		ctx.Error(apperrors.NewNotFound("课程不存在", nil))
		return
	}
	userID, _ := middleware.GetCurrentUserID(ctx)
	var uid *uint64
	if userID > 0 {
		uid = &userID
	}
	allowed, accessErr := c.accessChecker.canAccess(ctx, 2, courseID, uid, &course.AuthorID)
	if accessErr != nil {
		ctx.Error(accessErr)
		return
	}
	if !allowed {
		ctx.Error(apperrors.NewForbidden("无权限访问该课程单元附件", nil))
		return
	}
	fileIDs, listErr := c.unitAttachRepo.ListFileIDs(ctx, unitID)
	if listErr != nil {
		ctx.Error(listErr)
		return
	}
	response.Success(ctx, fileIDs)
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
