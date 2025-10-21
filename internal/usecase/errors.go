package usecase

import "errors"

// ErrAlreadyExist возвращается, если сущность уже существует.
var ErrAlreadyExist = errors.New("already exist")

// ErrCollision возвращается при неразрешимых коллизиях сгенерированных ссылок.
var ErrCollision = errors.New("unstopable collision")
