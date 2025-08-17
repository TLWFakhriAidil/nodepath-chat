package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
	
	"github.com/aldinokemal/go-whatsapp-web-multidevice/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserRepository handles user data persistence
type UserRepository struct {
	users       map[string]*models.User       // email -> user
	usersByID   map[string]*models.User       // id -> user
	devices     map[string][]*models.UserDevice // userID -> devices
	sessions    map[string]*models.UserSession  // token -> session
	mu          sync.RWMutex
	dataFile    string
	devicesFile string
}

var (
	userRepo     *UserRepository
	userRepoOnce sync.Once
)

// GetUserRepository returns singleton instance of UserRepository
func GetUserRepository() *UserRepository {
	userRepoOnce.Do(func() {
		userRepo = &UserRepository{
			users:       make(map[string]*models.User),
			usersByID:   make(map[string]*models.User),
			devices:     make(map[string][]*models.UserDevice),
			sessions:    make(map[string]*models.UserSession),
			dataFile:    "storages/users.json",
			devicesFile: "storages/user_devices.json",
		}
		userRepo.loadData()
		
		// Create default admin user if no users exist
		if len(userRepo.users) == 0 {
			userRepo.CreateUser("admin@whatsapp.com", "Administrator", "changeme123")
		}
	})
	return userRepo
}

// CreateUser creates a new user
func (r *UserRepository) CreateUser(email, fullName, password string) (*models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Check if user exists
	if _, exists := r.users[email]; exists {
		return nil, fmt.Errorf("user with email %s already exists", email)
	}
	
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	
	// Create user
	user := &models.User{
		ID:           uuid.New().String(),
		Email:        email,
		FullName:     fullName,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}
	
	r.users[email] = user
	r.usersByID[user.ID] = user
	r.saveData()
	
	return user, nil
}