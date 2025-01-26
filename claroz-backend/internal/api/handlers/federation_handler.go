package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lukelittle/claroz/claroz-backend/internal/federation"
	"github.com/lukelittle/claroz/claroz-backend/internal/models"
	"github.com/lukelittle/claroz/claroz-backend/internal/repository"
)

type FederationHandler struct {
	userRepo  repository.UserRepositoryInterface
	atpClient federation.ATProtoClientInterface
}

func NewFederationHandler(userRepo repository.UserRepositoryInterface, atpClient federation.ATProtoClientInterface) *FederationHandler {
	return &FederationHandler{
		userRepo:  userRepo,
		atpClient: atpClient,
	}
}

// ResolveRemoteProfile godoc
// @Summary Resolve a remote profile by handle
// @Description Resolves a remote profile from a federated handle (e.g. user.bsky.social)
// @Tags federation
// @Accept json
// @Produce json
// @Param handle path string true "Remote handle to resolve"
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /federation/resolve/{handle} [get]
func (h *FederationHandler) ResolveRemoteProfile(c *gin.Context) {
	handle := c.Param("handle")[1:] // Remove leading slash from wildcard
	if handle == "" || len(handle) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "handle is required and must be at least 3 characters"})
		return
	}

	// Check if we already have this profile
	existingUser, err := h.userRepo.FindByHandle(handle)
	if err != nil {
		// Only return error if it's not a "not found" error
		if existingUser != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check existing profile"})
			return
		}
	} else if existingUser != nil {
		c.JSON(http.StatusOK, existingUser)
		return
	}

	// Resolve remote profile
	fedProfile, err := h.atpClient.ResolveHandle(c.Request.Context(), handle)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "failed to resolve remote profile"})
		return
	}

	// Create local user record for remote profile
	user := &models.User{
		Handle:         fedProfile.Handle,
		DID:            fedProfile.DID,
		FullName:       fedProfile.DisplayName,
		Bio:            fedProfile.Description,
		Avatar:         fedProfile.Avatar,
		FederationType: "remote",
	}

	if err := h.userRepo.Create(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create local user record"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// SyncRemoteProfile godoc
// @Summary Sync a remote profile
// @Description Syncs the latest data for a remote profile
// @Tags federation
// @Accept json
// @Produce json
// @Param did path string true "DID of the remote profile"
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /federation/sync/{did} [post]
func (h *FederationHandler) SyncRemoteProfile(c *gin.Context) {
	did := c.Param("did")[1:] // Remove leading slash from wildcard
	if did == "" || len(did) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "did is required and must be at least 3 characters"})
		return
	}

	// Get existing user
	user, err := h.userRepo.FindByDID(did)
	if err != nil {
		if user == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check existing profile"})
		return
	}

	// Fetch latest remote profile
	fedProfile, err := h.atpClient.GetProfile(c.Request.Context(), did)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch remote profile"})
		return
	}

	// Update local record
	user.Handle = fedProfile.Handle
	user.FullName = fedProfile.DisplayName
	user.Bio = fedProfile.Description
	user.Avatar = fedProfile.Avatar

	if err := h.userRepo.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update local record"})
		return
	}

	c.JSON(http.StatusOK, user)
}
