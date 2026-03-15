# Domain Model

## Core Aggregates
- User
- Session
- UserProfile
- EquipmentCatalogItem
- ExerciseCatalogItem
- PlanVersion
- NutritionPlanVersion
- WorkoutScheduleInstance
- WorkoutLog
- MealLog
- WaterLog
- WeeklyCheckin
- Notification
- SupportThread
- DiscussionThread

## Important Rules
- Exactly one active plan version per user.
- Historical plan versions are immutable after supersession.
- Workout logs remain linked to the version active at log creation time.
- Reschedules create relations instead of replacing history.
- Deterministic targets are always backend-calculated.
- AI may only produce plan structure and localized user-facing content within the validated schema.

## Derived Concepts
- Adherence score.
- Plan health: `healthy`, `attention_needed`, `adaptation_recommended`.
- Regeneration trigger evaluation.
- Hydration adherence summary.

## Roles
- `user`
- `trainer`
- `admin`
