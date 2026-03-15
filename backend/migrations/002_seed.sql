INSERT INTO roles (id, code) VALUES
  ('00000000-0000-0000-0000-000000000001', 'user'),
  ('00000000-0000-0000-0000-000000000002', 'trainer'),
  ('00000000-0000-0000-0000-000000000003', 'admin')
ON CONFLICT DO NOTHING;

INSERT INTO equipment_catalog (id, slug, category, location_type, media_url, active) VALUES
  ('10000000-0000-0000-0000-000000000001', 'dumbbells', 'weights', 'mixed', 'https://example.com/media/dumbbells.jpg', TRUE),
  ('10000000-0000-0000-0000-000000000002', 'yoga-mat', 'accessory', 'home', 'https://example.com/media/mat.jpg', TRUE),
  ('10000000-0000-0000-0000-000000000003', 'barbell', 'weights', 'gym', 'https://example.com/media/barbell.jpg', TRUE)
ON CONFLICT DO NOTHING;

INSERT INTO equipment_catalog_translations (equipment_id, locale, name, description) VALUES
  ('10000000-0000-0000-0000-000000000001', 'ru', 'Гантели', 'Пара свободных весов'),
  ('10000000-0000-0000-0000-000000000001', 'kk', 'Gantel', 'Erkin salmaq jup'),
  ('10000000-0000-0000-0000-000000000002', 'ru', 'Коврик', 'Коврик для упражнений'),
  ('10000000-0000-0000-0000-000000000002', 'kk', 'Tosekse', 'Zhattyguga arnalgan tosekse'),
  ('10000000-0000-0000-0000-000000000003', 'ru', 'Штанга', 'Базовый силовой инвентарь'),
  ('10000000-0000-0000-0000-000000000003', 'kk', 'Shtanga', 'Negizgi kushi quraly')
ON CONFLICT DO NOTHING;

INSERT INTO exercise_catalog (id, slug, movement_pattern, difficulty, location_type, media_url, active) VALUES
  ('20000000-0000-0000-0000-000000000001', 'goblet-squat', 'squat', 'beginner', 'mixed', 'https://example.com/media/goblet-squat.jpg', TRUE),
  ('20000000-0000-0000-0000-000000000002', 'push-up', 'push', 'beginner', 'home', 'https://example.com/media/push-up.jpg', TRUE),
  ('20000000-0000-0000-0000-000000000003', 'dumbbell-row', 'pull', 'beginner', 'mixed', 'https://example.com/media/row.jpg', TRUE)
ON CONFLICT DO NOTHING;

INSERT INTO exercise_catalog_translations (exercise_id, locale, name, description, technique_steps) VALUES
  ('20000000-0000-0000-0000-000000000001', 'ru', 'Присед с гантелью', 'Базовое упражнение на ноги', 'Держи спину ровно; Сохраняй контроль; Дыши ровно'),
  ('20000000-0000-0000-0000-000000000001', 'kk', 'Gantelmen otyru', 'Aiaq ushin negizgi zhattygu', 'Arqany tike ustau; Qozgalysty baqylau; Tynysty birqalypty ustau'),
  ('20000000-0000-0000-0000-000000000002', 'ru', 'Отжимания', 'Базовое упражнение на верх тела', 'Корпус прямой; Локти под контролем; Не спеши'),
  ('20000000-0000-0000-0000-000000000002', 'kk', 'Iterilu', 'Denenin jogargy boligi ushin negizgi zhattygu', 'Deneni tike ustau; Qol bagytyn baqylau; Asyqpau'),
  ('20000000-0000-0000-0000-000000000003', 'ru', 'Тяга гантели в наклоне', 'Базовое упражнение на спину', 'Спина нейтральна; Тяни локтем; Контролируй опускание'),
  ('20000000-0000-0000-0000-000000000003', 'kk', 'Enkeide gantel tartu', 'Arqa ushin negizgi zhattygu', 'Arqa beitarap; Suyreudi shyntaqpen iste; Tusirudi baqyla')
ON CONFLICT DO NOTHING;
