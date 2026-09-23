BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

DELETE FROM users WHERE username IN ('Leszmak', 'Gulas', 'Sebas', 'Miłas');

INSERT INTO users (id, username, password_hash, color) VALUES
    (gen_random_uuid(), 'Leszmak',  crypt('123123', gen_salt('bf', 10)), '#e6194b'),
    (gen_random_uuid(), 'Gulas',    crypt('123123', gen_salt('bf', 10)), '#3cb44b'),
    (gen_random_uuid(), 'Sebas',    crypt('123123', gen_salt('bf', 10)), '#911eb4'),
    (gen_random_uuid(), 'Miłas',    crypt('123123', gen_salt('bf', 10)), '#f58231');

INSERT INTO activities (id, name, description, created_by)
SELECT gen_random_uuid(), a.name, a.description, u.id
FROM (VALUES
    ('Bieganie',   'Poranne i wieczorne rundki po parku', 'Leszmak'),
    ('Czytanie',   'Minimum 20 stron dziennie',           'Sebas'),
    ('Angielski',  'Duolingo albo lekcja z lektorem',     'Gulas'),
    ('Medytacja',  '10 minut ciszy',                      'Miłas'),
    ('Bez cukru',  'Dzień bez słodyczy i słodzonych napojów', 'Sebas')
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
    ('Leszmak',  'Gym',       'Siłownia PN/ŚR/PT',   TRUE,  FALSE, TRUE,  FALSE, TRUE,  FALSE, FALSE),
    ('Gulas',    'Gym',       'Trening wt/czw/sob',  FALSE, TRUE,  FALSE, TRUE,  FALSE, TRUE,  FALSE),
    ('Leszmak',  'Bieganie',  'Długie wybieganie',   FALSE, FALSE, FALSE, FALSE, FALSE, FALSE, TRUE),
    ('Miłas',    'Angielski', 'Lekcja z lektorem',   FALSE, TRUE,  FALSE, TRUE,  FALSE, FALSE, FALSE),
    ('Sebas',    'Bieganie',  'Truchtanie weekendowe', FALSE, FALSE, FALSE, FALSE, FALSE, TRUE, TRUE)
) AS p(username, activity, title, monday, tuesday, wednesday, thursday, friday, saturday, sunday)
JOIN users u ON u.username = p.username
JOIN activities a ON a.name = p.activity;

INSERT INTO plans (id, activity_id, user_id, title, monday, tuesday, wednesday, thursday, friday, saturday, sunday, starts_on)
SELECT id, activity_id, user_id, title, monday, tuesday, wednesday, thursday, friday, saturday, sunday,
       CURRENT_DATE - 120
FROM seed_plans;

INSERT INTO entries (id, activity_id, user_id, plan_id, scheduled_for, excused, description, occurred_at)
SELECT gen_random_uuid(), p.activity_id, p.user_id, p.id, d::date, r >= 0.75,
       CASE WHEN r >= 0.75
            THEN (ARRAY['Chory', 'Wyjazd służbowy', 'Kontuzja kolana', 'Urodziny babci'])[1 + floor(random() * 4)::int]
            ELSE pg_temp.seed_description(p.activity)
       END,
       d + time '07:00' + random() * interval '12 hours'
FROM seed_plans p
CROSS JOIN generate_series(CURRENT_DATE - 120, CURRENT_DATE, interval '1 day') AS d
CROSS JOIN LATERAL (SELECT random() AS r WHERE d IS NOT NULL) AS roll
WHERE r < 0.85
  AND (ARRAY[p.sunday, p.monday, p.tuesday, p.wednesday, p.thursday, p.friday, p.saturday])[extract(dow FROM d)::int + 1];

INSERT INTO entries (id, activity_id, user_id, description, occurred_at)
SELECT gen_random_uuid(), a.id, u.id,
       pg_temp.seed_description(a.name),
       d + time '06:00' + random() * interval '16 hours'
FROM (VALUES
    ('Leszmak',  'Bieganie',  0.15),
    ('Leszmak',  'Czytanie',  0.08),
    ('Gulas',    'Angielski', 0.30),
    ('Gulas',    'Bez cukru', 0.12),
    ('Sebas',    'Czytanie',  0.30),
    ('Sebas',    'Medytacja', 0.10),
    ('Sebas',    'Bez cukru', 0.15),
    ('Miłas',    'Medytacja', 0.20),
    ('Miłas',    'Gym',       0.05),
    ('Miłas',    'Czytanie',  0.10)
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
WHERE u.username IN ('Leszmak', 'Gulas', 'Sebas', 'Miłas')
  AND e.plan_id IS NULL
  AND random() + 0 * extract(epoch FROM e.occurred_at) < 0.08;

UPDATE entries e
SET occurred_at = LEAST(e.occurred_at, now() - random() * interval '1 hour')
FROM users u
WHERE u.id = e.user_id AND u.username IN ('Leszmak', 'Gulas', 'Sebas', 'Miłas');

UPDATE entries e
SET created_at = LEAST(e.occurred_at + random() * interval '36 hours', now())
FROM users u
WHERE u.id = e.user_id AND u.username IN ('Leszmak', 'Gulas', 'Sebas', 'Miłas');

COMMIT;

SELECT u.username, count(e.*) AS entries
FROM users u LEFT JOIN entries e ON e.user_id = u.id
GROUP BY u.username ORDER BY u.username;
