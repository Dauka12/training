package app

const developmentSeedPassword = "DevPassw0rd!123"
const bootstrapAdminPassword = "admin"

func (a *App) seedDevelopmentUsers() {
	if !isDevelopmentMode() {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	seeded := false
	seeded = a.ensureDevelopmentUserLockedWithPassword("admin@example.com", []string{"admin"}, bootstrapAdminPassword, true) || seeded
	seeded = a.ensureDevelopmentUserLocked("admin@local.test", []string{"admin"}) || seeded
	seeded = a.ensureDevelopmentUserLocked("trainer@local.test", []string{"trainer"}) || seeded
	seeded = a.ensureDevelopmentUserLocked("member@local.test", []string{"user"}) || seeded
	if seeded {
		a.persistStateLocked()
	}
}

func (a *App) ensureDevelopmentUserLocked(email string, roles []string) bool {
	return a.ensureDevelopmentUserLockedWithPassword(email, roles, developmentSeedPassword, false)
}

func (a *App) ensureDevelopmentUserLockedWithPassword(email string, roles []string, password string, mustChangePassword bool) bool {
	if _, exists := a.byMail[email]; exists {
		return false
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		a.log.Error("failed to hash development seed password", "email", email, "error", err)
		return false
	}

	user := &User{
		ID:                 token(8),
		Email:              email,
		PasswordHash:       passwordHash,
		Verified:           true,
		Locale:             "ru",
		Theme:              "light",
		Roles:              append([]string(nil), roles...),
		MustChangePassword: mustChangePassword,
		NotificationPreferences: NotificationPreferences{
			HydrationReminder: true,
		},
	}

	a.users[user.ID] = user
	a.byMail[user.Email] = user
	return true
}
