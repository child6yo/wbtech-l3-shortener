package repository

import "errors"

// ErrAlreadyExist возвращается, если сущность уже существует.
var ErrAlreadyExist = errors.New("already exist")
