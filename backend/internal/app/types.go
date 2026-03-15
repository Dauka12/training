package app

type User struct {
	ID                      string                  `json:"id"`
	Email                   string                  `json:"email"`
	PasswordHash            string                  `json:"password_hash"`
	AuthProvider            string                  `json:"auth_provider,omitempty"`
	Verified                bool                    `json:"verified"`
	Locale                  string                  `json:"locale"`
	Theme                   string                  `json:"theme"`
	Roles                   []string                `json:"roles"`
	SessionTokenHash        string                  `json:"session_token_hash"`
	CSRFTokenHash           string                  `json:"csrf_token_hash"`
	VerifyTokenHash         string                  `json:"verify_token_hash"`
	ResetTokenHash          string                  `json:"reset_token_hash"`
	VerifyToken             string                  `json:"-"`
	ResetToken              string                  `json:"-"`
	MustChangePassword      bool                    `json:"must_change_password"`
	OnboardingDone          bool                    `json:"onboarding_done"`
	Profile                 UserProfile             `json:"profile"`
	WaterTargetML           int                     `json:"water_target_ml"`
	WaterConsumed           int                     `json:"water_consumed"`
	WaterOverrideML         int                     `json:"water_override_ml"`
	PlanVersions            []PlanVersion           `json:"plan_versions"`
	Notifications           []Notification          `json:"notifications"`
	NotificationPreferences NotificationPreferences `json:"notification_preferences"`
	MealLogs                []MealLog               `json:"meal_logs"`
	WorkoutLogs             []WorkoutLog            `json:"workout_logs"`
	WaterLogs               []WaterLog              `json:"water_logs"`
	WeeklyCheckins          []WeeklyCheckin         `json:"weekly_checkins"`
	AIGenerationLogs        []AIGenerationLog       `json:"ai_generation_logs"`
	EmailLogs               []EmailLog              `json:"email_logs"`
	LastHydrationReminderAt string                  `json:"last_hydration_reminder_at,omitempty"`
	AssignedTrainerEmail    string                  `json:"assigned_trainer_email"`
}

type UserProfile struct {
	Age                   int               `json:"age"`
	BiologicalSex         string            `json:"biological_sex"`
	HeightCM              int               `json:"height_cm"`
	CurrentWeightKG       int               `json:"current_weight_kg"`
	TargetWeightKG        int               `json:"target_weight_kg"`
	PrimaryGoal           string            `json:"primary_goal"`
	ProgramDurationWeeks  int               `json:"program_duration_weeks"`
	ExperienceLevel       string            `json:"experience_level"`
	ActivityLevel         string            `json:"activity_level"`
	TrainingLocation      string            `json:"training_location"`
	Timezone              string            `json:"timezone"`
	PreferredTrainingStyle string           `json:"preferred_training_style"`
	PreferredMealStyle     string           `json:"preferred_meal_style"`
	HydrationPreference    string           `json:"hydration_preference"`
	AvailableTrainingDays []AvailabilityDay `json:"available_training_days"`
	EquipmentIDs          []string          `json:"equipment_ids"`
	Injuries              []string          `json:"injuries"`
	DietaryPreferences    []string          `json:"dietary_preferences"`
	AvoidFoods            []string          `json:"avoid_foods"`
}

type AvailabilityDay struct {
	Weekday     string `json:"weekday"`
	DurationMin int    `json:"duration_min"`
}

type PlanVersion struct {
	ID                 string          `json:"id"`
	VersionNo          int             `json:"version_no"`
	ParentPlanID       string          `json:"parent_plan_id,omitempty"`
	RegenerationReason string          `json:"regeneration_reason,omitempty"`
	SupersededAt       string          `json:"superseded_at,omitempty"`
	CreatedAt          string          `json:"created_at"`
	Title              string          `json:"title"`
	Summary            string          `json:"summary"`
	Nutrition          NutritionTarget `json:"nutrition"`
	Schedule           []ScheduleItem  `json:"schedule"`
	Weeks              []PlanWeek      `json:"weeks"`
	Warnings           []string        `json:"warnings"`
	AdaptationRules    []string        `json:"adaptation_rules"`
}

type NutritionTarget struct {
	DailyCalories int           `json:"daily_calories"`
	ProteinG      int           `json:"protein_g"`
	CarbsG        int           `json:"carbs_g"`
	FatG          int           `json:"fat_g"`
	DailyWaterML  int           `json:"daily_water_ml"`
	TrainingNote  string        `json:"training_note"`
	RestNote      string        `json:"rest_note"`
	HydrationNote string        `json:"hydration_note"`
	MealExamples  []MealExample `json:"meal_examples"`
}

type MealExample struct {
	Slot     string   `json:"slot"`
	Examples []string `json:"examples"`
}

type PlanWeek struct {
	WeekIndex int       `json:"week_index"`
	Days      []PlanDay `json:"days"`
}

type PlanDay struct {
	Weekday          string         `json:"weekday"`
	SessionName      string         `json:"session_name"`
	Goal             string         `json:"goal"`
	EstimatedMinutes int            `json:"estimated_minutes"`
	Warmup           []string       `json:"warmup"`
	Exercises        []PlanExercise `json:"exercises"`
	Cooldown         []string       `json:"cooldown"`
}

type PlanExercise struct {
	Order                   int      `json:"order"`
	ExerciseID              string   `json:"exercise_id"`
	ExerciseName            string   `json:"exercise_name"`
	Sets                    int      `json:"sets"`
	Reps                    string   `json:"reps"`
	RestSec                 int      `json:"rest_sec"`
	EffortNote              string   `json:"effort_note"`
	Notes                   string   `json:"notes"`
	SubstitutionExerciseIDs []string `json:"substitution_exercise_ids"`
	SubstitutionNames       []string `json:"substitution_names"`
}

type ScheduleItem struct {
	ID                string `json:"id"`
	Weekday           string `json:"weekday"`
	SessionName       string `json:"session_name"`
	EstimatedMinutes  int    `json:"estimated_minutes"`
	Status            string `json:"status"`
	RescheduledFromID string `json:"rescheduled_from_id,omitempty"`
}

type WorkoutLog struct {
	ScheduleID     string `json:"schedule_id"`
	Status         string `json:"status"`
	DiscomfortFlag bool   `json:"discomfort_flag"`
	Difficulty     int    `json:"difficulty"`
	Note           string `json:"note"`
	CompletionTime string `json:"completion_time"`
}

type MealLog struct {
	Status    string `json:"status"`
	Note      string `json:"note"`
	CreatedAt string `json:"created_at"`
}

type WaterLog struct {
	AmountML  int    `json:"amount_ml"`
	CreatedAt string `json:"created_at"`
}

type WeeklyCheckin struct {
	WeightKG            int    `json:"weight_kg"`
	EnergyLevel         int    `json:"energy_level"`
	AvailabilityChanged bool   `json:"availability_changed"`
	EquipmentChanged    bool   `json:"equipment_changed"`
	InjuryChanged       bool   `json:"injury_changed"`
	Note                string `json:"note"`
	CreatedAt           string `json:"created_at"`
}

type Notification struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Read      bool   `json:"read"`
	CreatedAt string `json:"created_at"`
	TargetURL string `json:"target_url,omitempty"`
}

type NotificationPreferences struct {
	HydrationReminder bool `json:"hydration_reminder"`
	EmailEnabled      bool `json:"email_enabled"`
}

type AIGenerationLog struct {
	ID          string `json:"id"`
	RequestedAt string `json:"requested_at"`
	Status      string `json:"status"`
	Provider    string `json:"provider"`
	PlanTitle   string `json:"plan_title,omitempty"`
	Error       string `json:"error,omitempty"`
	Attempt     int    `json:"attempt"`
}

type EmailLog struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	To        string `json:"to"`
	Subject   string `json:"subject"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
	CreatedAt string `json:"created_at"`
}

type AuditLog struct {
	ID           string `json:"id"`
	ActorEmail   string `json:"actor_email"`
	Action       string `json:"action"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	CreatedAt    string `json:"created_at"`
}

type TrainerNote struct {
	ID           string `json:"id"`
	TrainerEmail string `json:"trainer_email"`
	UserEmail    string `json:"user_email"`
	Body         string `json:"body"`
	CreatedAt    string `json:"created_at"`
}

type SupportThread struct {
	ID              string           `json:"id"`
	UserEmail       string           `json:"user_email"`
	Title           string           `json:"title"`
	Status          string           `json:"status"`
	AssignedToEmail string           `json:"assigned_to_email"`
	Messages        []SupportMessage `json:"messages"`
	CreatedAt       string           `json:"created_at"`
}

type SupportMessage struct {
	AuthorEmail string `json:"author_email"`
	Body        string `json:"body"`
	CreatedAt   string `json:"created_at"`
}

type DiscussionThread struct {
	ID          string            `json:"id"`
	AuthorEmail string            `json:"author_email"`
	Title       string            `json:"title"`
	Body        string            `json:"body"`
	Category    string            `json:"category"`
	Status      string            `json:"status"`
	Replies     []DiscussionReply `json:"replies"`
	CreatedAt   string            `json:"created_at"`
}

type DiscussionReply struct {
	AuthorEmail string `json:"author_email"`
	Body        string `json:"body"`
	CreatedAt   string `json:"created_at"`
}

type EquipmentItem struct {
	ID           string            `json:"id"`
	Names        map[string]string `json:"names"`
	Descriptions map[string]string `json:"descriptions"`
	Category     string            `json:"category"`
	LocationType string            `json:"location_type"`
	MediaURL     string            `json:"media_url"`
	Active       bool              `json:"active"`
}

type ExerciseItem struct {
	ID           string            `json:"id"`
	Slug         string            `json:"slug"`
	Names        map[string]string `json:"names"`
	Descriptions map[string]string `json:"descriptions"`
	Technique    map[string]string `json:"technique"`
	Movement     string            `json:"movement_pattern"`
	Difficulty   string            `json:"difficulty"`
	LocationType string            `json:"location_type"`
	EquipmentIDs []string          `json:"equipment_ids"`
	MediaURL     string            `json:"media_url"`
	ContraindicationTags []string  `json:"contraindication_tags"`
	SubstitutionIDs      []string  `json:"substitution_ids"`
	Active       bool              `json:"active"`
}

type AIProvider interface {
	GeneratePlan(input GenerationInput) (GeneratedPlan, error)
}

type GenerationInput struct {
	Locale            string
	UserRef           string
	Profile           UserProfile
	Targets           NutritionTarget
	Candidates        []ExerciseItem
	SelectedEquipment []EquipmentItem
	History           GenerationHistory
}

type GeneratedPlan struct {
	Title           string
	Summary         string
	Warnings        []string
	Nutrition       NutritionTarget
	Sessions        []ScheduleItem
	Weeks           []PlanWeek
	AdaptationRules []string
}

type GenerationHistory struct {
	PreviousPlanSummary          string
	CompletedSessionsLast14Days  int
	MissedSessionsLast14Days     int
	MealAdherenceLast14Days      int
	HydrationAdherenceLast14Days int
	LatestWeightKG               int
	ReasonForRegeneration        string
}
