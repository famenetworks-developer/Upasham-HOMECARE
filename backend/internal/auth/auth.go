package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type CustomClaims struct {
	UserID           string `json:"user_id"`
	Email            string `json:"email"`
	Role             string `json:"role"`
	FullName         string `json:"full_name"`
	PatientAccountID string `json:"patient_account_id,omitempty"`
	DoctorID         string `json:"doctor_id,omitempty"`
	jwt.RegisteredClaims
}

var jwtSecretKey []byte

func SetJWTSecret(secret string) {
	jwtSecretKey = []byte(secret)
}

func GenerateToken(userID, email, role, fullName, patientAccountID, doctorID string) (string, error) {
	claims := CustomClaims{
		UserID:           userID,
		Email:            email,
		Role:             role,
		FullName:         fullName,
		PatientAccountID: patientAccountID,
		DoctorID:         doctorID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "upasham-connected-care",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecretKey)
}

func ValidateToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func CheckPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// Middleware: RequireAuth extracts and verifies Bearer token
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must be Bearer <token>"})
			c.Abort()
			return
		}

		claims, err := ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("full_name", claims.FullName)
		c.Set("patient_account_id", claims.PatientAccountID)
		c.Set("doctor_id", claims.DoctorID)
		c.Next()
	}
}

// Middleware: RequireRole checks if the user has any of the allowed roles
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		userRole := roleVal.(string)
		allowed := false
		for _, r := range allowedRoles {
			if r == userRole || userRole == "admin" { // admin has universal supervisory access
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access forbidden: insufficient role permissions",
				"required_roles": allowedRoles,
				"your_role": userRole,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
