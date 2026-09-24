BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

DELETE FROM users WHERE username IN ('Ala', 'Bartek', 'Celina', 'Darek');

INSERT INTO users (id, username, password_hash, color) VALUES
    (gen_random_uuid(), 'Ala',  crypt('123123', gen_salt('bf', 10)), '#e6194b'),
    (gen_random_uuid(), 'Bartek',    crypt('123123', gen_salt('bf', 10)), '#3cb44b'),
    (gen_random_uuid(), 'Celina',    crypt('123123', gen_salt('bf', 10)), '#911eb4'),
    (gen_random_uuid(), 'Darek',    crypt('123123', gen_salt('bf', 10)), '#f58231');

INSERT INTO activities (id, name, description, created_by)
SELECT gen_random_uuid(), a.name, a.description, u.id
FROM (VALUES
    ('Gym',        'Siłownia, każdy trening się liczy',  'Ala'),
    ('Bieganie',   'Poranne i wieczorne rundki po parku', 'Ala'),
    ('Czytanie',   'Minimum 20 stron dziennie',           'Celina'),
    ('Angielski',  'Duolingo albo lekcja z lektorem',     'Bartek'),
    ('Medytacja',  '10 minut ciszy',                      'Darek'),
    ('Bez cukru',  'Dzień bez słodyczy i słodzonych napojów', 'Celina')
) AS a(name, description, username)
JOIN users u ON u.username = a.username
ON CONFLICT (name) DO NOTHING;

CREATE FUNCTION pg_temp.seed_description(activity TEXT) RETURNS TEXT AS $$
    SELECT options[1 + floor(random() * array_length(options, 1))::int]
    FROM (SELECT CASE activity
        WHEN 'Gym' THEN ARRAY['Klata i triceps', 'Dzień nóg, ledwo schodzę po schodach', 'Plecy i biceps', 'Martwy ciąg, nowy rekord!', 'Barki, krótko ale intensywnie', 'Full body z trenerem']
        WHEN 'Bieganie' THEN ARRAY['5 km wokół parku', 'Interwały 8x400 m', '10 km spokojnym tempem', 'Bieg w deszczu, było warto', 'Rozbieganie 3 km', 'Podbiegi na Kopcu']
        WHEN 'Czytanie' THEN ARRAY['30 stron Wiedźmina', 'Rozdział o nawykach', 'Reportaż do poduszki', 'Skończyłem książkę!', '25 stron w tramwaju', 'Kryminał, nie mogłem się oderwać']
        WHEN 'Angielski' THEN ARRAY['Lekcja Duolingo, seria trwa', 'Konwersacje z lektorem', 'Odcinek serialu bez napisów', 'Phrasal verbs, masakra', 'Fiszki: 40 nowych słówek', 'Podcast po angielsku']
        WHEN 'Medytacja' THEN ARRAY['10 minut z oddechem', 'Body scan przed snem', 'Rano, zanim wszyscy wstali', 'Ciężko było się skupić', '15 minut w ciszy', 'Medytacja z aplikacją']
        WHEN 'Bez cukru' THEN ARRAY['Odmówiłem ciastka w pracy', 'Bez słodyczy, tylko owoce', 'Herbata bez cukru, da się', 'Ominąłem automat z batonami', 'Urodziny w pracy, wytrwałem', 'Cały dzień czysto']
        ELSE ARRAY['Zrobione']
    END AS options) AS o
$$ LANGUAGE sql VOLATILE;

CREATE TEMP TABLE seed_plans ON COMMIT DROP AS
SELECT gen_random_uuid() AS id, a.id AS activity_id, u.id AS user_id, p.*
FROM (VALUES
    ('Ala',  'Gym',       'Siłownia PN/ŚR/PT',   TRUE,  FALSE, TRUE,  FALSE, TRUE,  FALSE, FALSE),
    ('Bartek',    'Gym',       'Trening wt/czw/sob',  FALSE, TRUE,  FALSE, TRUE,  FALSE, TRUE,  FALSE),
    ('Ala',  'Bieganie',  'Długie wybieganie',   FALSE, FALSE, FALSE, FALSE, FALSE, FALSE, TRUE),
    ('Darek',    'Angielski', 'Lekcja z lektorem',   FALSE, TRUE,  FALSE, TRUE,  FALSE, FALSE, FALSE),
    ('Celina',    'Bieganie',  'Truchtanie weekendowe', FALSE, FALSE, FALSE, FALSE, FALSE, TRUE, TRUE),
    ('Celina',    'Czytanie',  'Czytanie po pracy',   TRUE,  FALSE, TRUE,  FALSE, TRUE,  FALSE, FALSE),
    ('Bartek',    'Bez cukru', 'Tydzień bez cukru',   TRUE,  TRUE,  TRUE,  TRUE,  TRUE,  FALSE, FALSE),
    ('Darek',    'Medytacja', 'Medytacja co rano',   TRUE,  TRUE,  TRUE,  TRUE,  TRUE,  TRUE,  TRUE)
) AS p(username, activity, title, monday, tuesday, wednesday, thursday, friday, saturday, sunday)
JOIN users u ON u.username = p.username
JOIN activities a ON a.name = p.activity;

INSERT INTO plans (id, activity_id, user_id, title, monday, tuesday, wednesday, thursday, friday, saturday, sunday, starts_on)
SELECT id, activity_id, user_id, title, monday, tuesday, wednesday, thursday, friday, saturday, sunday,
       CURRENT_DATE - 120
FROM seed_plans;

CREATE TEMP TABLE seed_habits ON COMMIT DROP AS
SELECT * FROM (VALUES
    ('Ala', 0.97, 0.05, 0.97, 0.05),
    ('Bartek',   0.85, 0.65, 0.80, 0.80),
    ('Celina',   0.95, 0.05, 0.60, 0.60),
    ('Darek',   0.45, 0.20, 0.35, 0.30)
) AS h(username, good_show, good_excuse, bad_show, bad_excuse);

INSERT INTO entries (id, activity_id, user_id, plan_id, scheduled_for, excused, description, occurred_at)
SELECT gen_random_uuid(), p.activity_id, p.user_id, p.id, d::date, roll.excused,
       CASE WHEN roll.excused
            THEN (ARRAY['Chory', 'Wyjazd służbowy', 'Kontuzja kolana', 'Urodziny babci', 'Zaspałem', 'Pogoda nie ta'])[1 + floor(random() * 6)::int]
            ELSE pg_temp.seed_description(p.activity)
       END,
       (d::date + time '07:00' + random() * interval '12 hours') AT TIME ZONE 'Europe/Warsaw'
FROM seed_plans p
JOIN users u ON u.id = p.user_id
JOIN seed_habits h ON h.username = u.username
CROSS JOIN generate_series(CURRENT_DATE - 120, CURRENT_DATE - 1, interval '1 day') AS d
CROSS JOIN LATERAL (
    SELECT extract(week FROM d)::int % 2 = extract(week FROM CURRENT_DATE)::int % 2 AS good_week
) AS w
CROSS JOIN LATERAL (
    SELECT random() < CASE WHEN w.good_week THEN h.good_show ELSE h.bad_show END AS shows,
           random() < CASE WHEN w.good_week THEN h.good_excuse ELSE h.bad_excuse END AS excused
    WHERE d IS NOT NULL
) AS roll
WHERE roll.shows
  AND (ARRAY[p.sunday, p.monday, p.tuesday, p.wednesday, p.thursday, p.friday, p.saturday])[extract(dow FROM d)::int + 1];

INSERT INTO entries (id, activity_id, user_id, description, occurred_at)
SELECT gen_random_uuid(), a.id, u.id,
       pg_temp.seed_description(a.name),
       (d::date + time '06:00' + random() * interval '16 hours') AT TIME ZONE 'Europe/Warsaw'
FROM (VALUES
    ('Ala',  'Bieganie',  0.15),
    ('Ala',  'Czytanie',  0.08),
    ('Bartek',    'Angielski', 0.30),
    ('Bartek',    'Bez cukru', 0.12),
    ('Celina',    'Czytanie',  0.15),
    ('Celina',    'Medytacja', 0.10),
    ('Celina',    'Bez cukru', 0.15),
    ('Darek',    'Bieganie',  0.25),
    ('Darek',    'Gym',       0.05),
    ('Darek',    'Czytanie',  0.10)
) AS h(username, activity, chance)
JOIN users u ON u.username = h.username
JOIN activities a ON a.name = h.activity
CROSS JOIN generate_series(CURRENT_DATE - 120, CURRENT_DATE, interval '1 day') AS d
CROSS JOIN LATERAL (SELECT random() AS r WHERE d IS NOT NULL AND h.chance IS NOT NULL) AS roll
WHERE r < h.chance;

INSERT INTO entries (id, activity_id, user_id, description, occurred_at)
SELECT gen_random_uuid(), e.activity_id, e.user_id, 'Druga runda: ' || lower(pg_temp.seed_description(a.name)), e.occurred_at + interval '2 hours'
FROM entries e
JOIN activities a ON a.id = e.activity_id
JOIN users u ON u.id = e.user_id
WHERE u.username IN ('Ala', 'Bartek', 'Celina', 'Darek')
  AND e.plan_id IS NULL
  AND random() + 0 * extract(epoch FROM e.occurred_at) < 0.08;

UPDATE entries e
SET occurred_at = LEAST(e.occurred_at, now() - random() * interval '1 hour')
FROM users u
WHERE u.id = e.user_id AND u.username IN ('Ala', 'Bartek', 'Celina', 'Darek');

UPDATE entries e
SET created_at = LEAST(e.occurred_at + random() * interval '36 hours', now())
FROM users u
WHERE u.id = e.user_id AND u.username IN ('Ala', 'Bartek', 'Celina', 'Darek');

COMMIT;

SELECT u.username, count(e.*) AS entries
FROM users u LEFT JOIN entries e ON e.user_id = u.id
GROUP BY u.username ORDER BY u.username;
