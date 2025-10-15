package role

type Role string

const (
	Guest     Role = "guest"     // незарегистрированный пользователь
	Buyer     Role = "buyer"     // зарегистрированный пользователь
	Moderator Role = "moderator" // модератор
)

// FromUser определяет роль на основе данных пользователя
func FromUser(userID uint, isModerator bool) Role {
	if userID == 0 {
		return Guest
	}
	if isModerator {
		return Moderator
	}
	return Buyer
}

// HasAccess проверяет доступ для операций модератора
func (r Role) HasModeratorAccess() bool {
	return r == Moderator
}

// IsAuthenticated проверяет что пользователь авторизован
func (r Role) IsAuthenticated() bool {
	return r == Buyer || r == Moderator
}
