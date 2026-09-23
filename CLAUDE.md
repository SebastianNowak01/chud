# CLAUDE.md

- Bez komentarzy w kodzie.
- UI tylko w Tailwindzie; tokeny w `@theme` (`chud-ui/src/tailwind.css`), wspólne komponenty w `chud-ui/src/components/ui/`. Stylów drawably nie ruszać.
- Tekst dla użytkownika po polsku (UI i błędy API); błędy 500 mogą być po angielsku.
- Zdalna baza jest nietykalna: żadnych zapytań, zmian ani nowych baz bez wyraźnej zgody. Lokalną bazę dev (`chud-db`, `chud`) wolno nadpisywać przy testach; nie twórz obok niej nowych baz.

## Kontrakty modelu danych

Strukturę pilnuje baza (`schema.sql`), czas i harmonogram — serwisy w Go.

**Czas**
- Jedna strefa: `APP_TIMEZONE`, domyślnie `Europe/Warsaw` (`platform/clock`).
- Daty jako `YYYY-MM-DD`, liczone w tej strefie.

**Plany**
- ≥1 dzień tygodnia, niepusty tytuł, `ends_on >= starts_on`.
- Start nie w przeszłości.
- Dni i start niezmienne; edycja tylko tytułu i końca.
- Koniec najwcześniej wczoraj i nie przed ostatnim wpisem planu.
- Usunąć można tylko niezaczęty plan bez wpisów; resztę się kończy.
- Plany mogą się nakładać, każdy dzień liczy się osobno.

**Wpisy**
- Poza planem: bez `plan_id`/`scheduled_for`, nigdy wymówka, może każdy.
- Do planu: `plan_id` + `scheduled_for` razem, jeden na dzień, dzień z harmonogramu.
- Wpis do planu ma osobę i aktywność planu.
- Wymówka tylko w planie, zawsze z opisem.
- Brak wpisów z przyszłości.
- Zrobione tylko tego samego dnia; wymówka kiedykolwiek.

**Statusy dni** (liczone w backendzie, niezapisywane)
- zrobione / wymówka / opuszczone (przed dziś, brak wpisu) / do zrobienia.

**Ranking**
- Punkty: zrobione +3, poza planem +1, wymówka 0, opuszczone −2.
- Skuteczność = zrobione / (zrobione + opuszczone).
- Sortowanie: punkty, zrobione, nazwa.
