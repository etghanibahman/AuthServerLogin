package domain

import (
	"errors"
	"fmt"

	"bankingAuth/logger"

	"bankingAuth/errs"

	"gorm.io/gorm"
)

type AuthRepository interface {
	FindBy(username string, password string) (*Login, *errs.AppError)
	GenerateAndSaveRefreshTokenToStore(authToken AuthToken) (string, *errs.AppError)
	RefreshTokenExists(refreshToken string) *errs.AppError
}

type AuthRepositoryDb struct {
	client *gorm.DB
}

func (d AuthRepositoryDb) RefreshTokenExists(refreshToken string) *errs.AppError {
	sqlSelect := "select refresh_token from refresh_token_store where refresh_token = ?"
	var token string
	//err := d.client.Get(&token, sqlSelect, refreshToken)
	err := d.client.Raw(sqlSelect, refreshToken).Scan(&token).Error
	// if err != nil {
	// 	if err == sql.ErrNoRows {
	// 		return errs.NewAuthenticationError("refresh token not registered in the store")
	// 	} else {
	// 		logger.Error("Unexpected database error: " + err.Error())
	// 		return errs.NewUnexpectedError("unexpected database error")
	// 	}
	// }
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NewAuthenticationError("refresh token not registered in the store")
		} else {
			// Other error occurred
			logger.Error("Unexpected database error: " + err.Error())
			return errs.NewUnexpectedError("unexpected database error")
		}
	}
	return nil
}

func (d AuthRepositoryDb) GenerateAndSaveRefreshTokenToStore(authToken AuthToken) (string, *errs.AppError) {
	// generate the refresh token
	var appErr *errs.AppError
	var refreshToken string
	if refreshToken, appErr = authToken.newRefreshToken(); appErr != nil {
		return "", appErr
	}

	// store it in the store
	sqlInsert := "insert into refresh_token_store (refresh_token) values (?)"
	result := d.client.Exec(sqlInsert, refreshToken)
	if result.Error != nil {
		logger.Error("unexpected database error: " + result.Error.Error())
		return "", errs.NewUnexpectedError("unexpected database error")
	}
	return refreshToken, nil
}

func (d AuthRepositoryDb) FindBy(username, password string) (*Login, *errs.AppError) {
	fmt.Printf("Find by is called")
	var login Login
	sqlVerify := `SELECT u.username, u.customer_id, u.role, STRING_AGG(a.account_id::TEXT, ',') AS account_numbers
	FROM users u
	LEFT JOIN accounts a ON a.customer_id = u.customer_id
	WHERE u.username = ? AND u.password = ?
	GROUP BY u.username, u.customer_id, u.role;`
	//err := d.client.Get(&login, sqlVerify, username, password)
	err := d.client.Raw(sqlVerify, username, password).Scan(&login).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Record not found
			return nil, errs.NewAuthenticationError("invalid credentials")
		} else {
			// Other error occurred
			logger.Error("Error while verifying login request from database: " + err.Error())
			//logger.Error("Error while verifying login request from database: "  )
			return nil, errs.NewUnexpectedError("Unexpected database error")
		}
	}
	return &login, nil
}

func NewAuthRepository(client *gorm.DB) AuthRepositoryDb {
	return AuthRepositoryDb{client}
}
