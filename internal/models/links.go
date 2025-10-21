package models

// ShortLink сокращенная ссылка на ресурс, кастомная или генерируемая сервисом.
// Существует для редиректа на FullLink.
type ShortLink string

// FullLink полная ссылка, ведующая непосредственно на ресурс.
type FullLink string

// Link представляет модель ссылок.
type Link struct {
	Short ShortLink
	Full  FullLink
}
