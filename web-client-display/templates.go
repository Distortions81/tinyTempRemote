package main

const dashboardTemplateHTML = `<!DOCTYPE html>
<html>
<head>
<title>Weather & Time</title>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
:root {
  --background: {{ .Theme.Background }};
  --text-primary: {{ .Theme.TextPrimary }};
  --text-secondary: {{ .Theme.TextSecondary }};
  --text-hint: {{ .Theme.TextHint }};
  --accent: {{ .Theme.Accent }};
  --card-background: {{ .Theme.CardBackground }};
  --card-text: {{ .Theme.CardText }};
  --card-subtext: {{ .Theme.CardSubtext }};
  --card-border: {{ .Theme.CardBorder }};
  --card-shadow: {{ .Theme.CardShadow }};
}
* {
  box-sizing: border-box;
}
body {
  background: var(--background);
  color: var(--text-primary);
  font-family: "Segoe UI", sans-serif;
  margin: 0;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  min-height: 100vh;
  padding: 1.5rem 1rem;
}
.content {
  width: min(90vw, 760px);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.8rem;
  text-align: center;
}
h1 {
  margin: 0;
  font-size: clamp(2rem, 4vw, 3rem);
  letter-spacing: 0.08em;
  color: var(--text-primary);
}
.clock {
  font-size: clamp(3rem, 10vw, 5rem);
  letter-spacing: 0.15em;
  font-weight: 600;
  color: var(--accent);
}
.temp {
  font-size: clamp(2.4rem, 7vw, 4rem);
  font-weight: 500;
  color: var(--accent);
}
.desc {
  color: var(--text-secondary);
  font-size: clamp(1.1rem, 3.2vw, 1.6rem);
  margin-bottom: 0.6rem;
}
.details {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 0.6rem;
  margin-bottom: 0.6rem;
}
.detail-item {
  background: var(--card-background);
  border: 1px solid var(--card-border);
  border-radius: 14px;
  padding: 0.55rem 0.9rem;
  min-width: 150px;
  box-shadow: var(--card-shadow);
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}
.detail-label {
  font-size: 0.82rem;
  color: var(--card-subtext);
  letter-spacing: 0.06em;
  text-transform: uppercase;
}
.detail-value {
  font-size: 1.08rem;
  font-weight: 600;
  color: var(--card-text);
}
.hint {
  color: var(--text-hint);
  font-size: clamp(0.95rem, 2.5vw, 1.2rem);
  max-width: 520px;
}
.settings-link {
  align-self: flex-end;
  color: var(--text-secondary);
  text-decoration: none;
  font-size: 0.95rem;
  padding: 0.35rem 0.75rem;
  border-radius: 999px;
  border: 1px solid var(--card-border);
  background: var(--card-background);
  box-shadow: var(--card-shadow);
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}
.settings-link:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 12px rgba(0, 0, 0, 0.12);
}
.forecast {
  display: flex;
  justify-content: center;
  gap: 0.8rem;
  flex-wrap: wrap;
  width: 100%;
}
.day {
  background: var(--card-background);
  border-radius: 16px;
  padding: 0.75rem;
  flex: 1 1 110px;
  max-width: 140px;
  min-width: 110px;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  border: 1px solid var(--card-border);
  box-shadow: var(--card-shadow);
  color: var(--card-text);
}
.day h3 {
  margin: 0;
  font-size: clamp(1.1rem, 2.8vw, 1.4rem);
  letter-spacing: 0.05em;
  color: var(--card-text);
}
.day-temp,
.day-desc {
  margin: 0;
}
.day-temp {
  font-size: clamp(1rem, 2.5vw, 1.3rem);
  color: var(--accent);
  font-weight: 600;
}
.day-desc {
  font-size: clamp(0.82rem, 2vw, 1rem);
  color: var(--card-subtext);
}
.day-details {
  list-style: none;
  padding: 0;
  margin: 0.35rem 0 0;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  color: var(--card-subtext);
  font-size: 0.9rem;
}
.day-details li {
  display: flex;
  justify-content: space-between;
  gap: 0.4rem;
}
.day-details span:last-child {
  color: var(--card-text);
  font-weight: 600;
}
</style>
<script>
const TIME_FORMAT = "{{ .TimeFormat }}";
const SHOW_SECONDS = {{ if .ShowSeconds }}true{{ else }}false{{ end }};

function updateClock() {
  const now = new Date();
  let hours = now.getHours();
  const minutes = now.getMinutes();
  const seconds = now.getSeconds();

  if (TIME_FORMAT === '24h') {
    let timeText =
      hours.toString().padStart(2,'0') + ':' +
      minutes.toString().padStart(2,'0');
    if (SHOW_SECONDS) {
      timeText += ':' + seconds.toString().padStart(2,'0');
    }
    document.getElementById('clock').innerText = timeText;
    return;
  }

  const ampm = hours >= 12 ? 'PM' : 'AM';
  hours = hours % 12;
  hours = hours ? hours : 12;
  let timeText =
    hours.toString().padStart(2,'0') + ':' +
    minutes.toString().padStart(2,'0');
  if (SHOW_SECONDS) {
    timeText += ':' + seconds.toString().padStart(2,'0');
  }
  timeText += ' ' + ampm;
  document.getElementById('clock').innerText = timeText;
}

setInterval(updateClock, 1000);
window.onload = updateClock;
</script>
</head>
<body>
  <div class="content">
    <a class="settings-link" href="{{ .SettingsURL }}">Settings</a>
    {{ if .Error }}
      <h1>Weather data unavailable</h1>
    {{ else if .City }}
      <h1>{{ .City }}</h1>
    {{ end }}
    <div class="clock" id="clock"></div>
    {{ if .Error }}
      <div class="desc">{{ .Error }}</div>
      {{ if .Hint }}
        <p class="hint">{{ .Hint }}</p>
      {{ end }}
    {{ else }}
      <div class="temp">{{ .Temp }}°{{ .Unit }}</div>
      <div class="desc">{{ .Desc }}</div>
      {{ if .TodayDetails }}
        <div class="details">
          {{ range .TodayDetails }}
            <div class="detail-item">
              <span class="detail-label">{{ .Label }}</span>
              <span class="detail-value">{{ .Value }}</span>
            </div>
          {{ end }}
        </div>
      {{ end }}
      {{ if .Forecast }}
        <div class="forecast">
          {{ range .Forecast }}
            <div class="day">
              <h3>{{ .Day }}</h3>
              <p class="day-temp">{{ .Temp }}°{{ $.Unit }}</p>
              <p class="day-desc">{{ .Desc }}</p>
              {{ if .Details }}
                <ul class="day-details">
                  {{ range .Details }}
                    <li>
                      <span>{{ .Label }}</span>
                      <span>{{ .Value }}</span>
                    </li>
                  {{ end }}
                </ul>
              {{ end }}
            </div>
          {{ end }}
        </div>
      {{ else if .ForecastError }}
        <p class="hint">{{ .ForecastError }}</p>
      {{ end }}
    {{ end }}
  </div>
</body>
</html>`

const settingsTemplateHTML = `<!DOCTYPE html>
<html>
<head>
<title>piWeather Settings</title>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
* {
  box-sizing: border-box;
}
body {
  margin: 0;
  font-family: "Segoe UI", sans-serif;
  background: #0c1a2a;
  color: #f5f9ff;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  min-height: 100vh;
  padding: 1.5rem 1rem;
}
.container {
  width: min(96vw, 660px);
  background: rgba(10, 22, 34, 0.92);
  border-radius: 18px;
  padding: 1.8rem 1.6rem;
  box-shadow: 0 20px 45px rgba(0, 0, 0, 0.35);
  border: 1px solid rgba(255, 255, 255, 0.08);
}
.back-link {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  color: #9cc2ff;
  text-decoration: none;
  font-size: 0.95rem;
}
.back-link span {
  font-size: 1.2rem;
}
.back-link:hover {
  text-decoration: underline;
}
h1 {
  margin: 0.6rem 0 1.5rem;
  font-size: clamp(1.8rem, 4vw, 2.4rem);
  letter-spacing: 0.05em;
}
.status {
  margin: 0 0 1rem;
  padding: 0.6rem 0.8rem;
  border-radius: 10px;
  background: rgba(46, 204, 113, 0.2);
  color: #2ecc71;
}
.error {
  margin: 0 0 1rem;
  padding: 0.6rem 0.8rem;
  border-radius: 10px;
  background: rgba(231, 76, 60, 0.2);
  color: #ff8579;
}
.form-group {
  margin-bottom: 1.1rem;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}
label {
  font-weight: 600;
  letter-spacing: 0.04em;
  font-size: 0.95rem;
}
input[type="text"],
input[type="number"],
select {
  padding: 0.65rem 0.75rem;
  border-radius: 10px;
  border: 1px solid rgba(255, 255, 255, 0.15);
  background: rgba(7, 17, 27, 0.85);
  color: #f5f9ff;
  font-size: 1rem;
}
input[type="text"]::placeholder,
input[type="number"]::placeholder {
  color: rgba(255, 255, 255, 0.45);
}
.help {
  color: rgba(255, 255, 255, 0.6);
  font-size: 0.9rem;
  margin: -0.2rem 0 0;
}
.radio-group {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
}
.radio {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.45rem 0.7rem;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid transparent;
  cursor: pointer;
  font-size: 0.95rem;
}
.radio input {
  accent-color: #4da3ff;
}
.radio input:checked + span {
  font-weight: 600;
}
.radio:hover {
  border-color: rgba(255, 255, 255, 0.2);
}
.detail-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(190px, 1fr));
  gap: 0.75rem;
}
.detail-setting {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  padding: 0.6rem 0.7rem;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.1);
}
.detail-setting span {
  font-size: 0.9rem;
  font-weight: 600;
  letter-spacing: 0.04em;
  color: rgba(255, 255, 255, 0.8);
}
button {
  width: 100%;
  padding: 0.75rem;
  font-size: 1.05rem;
  letter-spacing: 0.08em;
  border-radius: 12px;
  border: none;
  background: linear-gradient(135deg, #4b92ff, #5ad8ff);
  color: #0b121d;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}
button:hover {
  transform: translateY(-1px);
  box-shadow: 0 12px 25px rgba(90, 216, 255, 0.35);
}
a {
  color: #9cc2ff;
}
</style>
</head>
<body>
  <div class="container">
    <a class="back-link" href="{{ .DashboardURL }}"><span>←</span>Back to dashboard</a>
    <h1>Dashboard Settings</h1>
    {{ if .Status }}
      <p class="status">{{ .Status }}</p>
    {{ end }}
    {{ if .Error }}
      <p class="error">{{ .Error }}</p>
    {{ end }}
    <form method="post" class="settings-form">
      <div class="form-group">
        <label for="api_key">OpenWeather API key</label>
        <input
          id="api_key"
          name="openweather_api_key"
          type="text"
          value="{{ .FormValues.Get "openweather_api_key" }}"
          required
        >
        <p class="help">Create or manage keys at <a href="https://home.openweathermap.org/api_keys" target="_blank" rel="noopener">OpenWeather</a>.</p>
      </div>
      <div class="form-group">
        <label for="location_name">Location label (optional)</label>
        <input
          id="location_name"
          name="location_name"
          type="text"
          value="{{ .FormValues.Get "location_name" }}"
          placeholder="Displayed name on the dashboard"
        >
      </div>
      <div class="form-group">
        <label>Coordinates</label>
        <div style="display: flex; gap: 0.6rem; flex-wrap: wrap;">
          <div style="flex: 1 1 180px;">
            <input
              id="latitude"
              name="latitude"
              type="number"
              step="0.0001"
              value="{{ .FormValues.Get "latitude" }}"
              placeholder="Latitude"
              required
            >
          </div>
          <div style="flex: 1 1 180px;">
            <input
              id="longitude"
              name="longitude"
              type="number"
              step="0.0001"
              value="{{ .FormValues.Get "longitude" }}"
              placeholder="Longitude"
              required
            >
          </div>
        </div>
      </div>
      <div class="form-group">
        <label>Units</label>
        <div class="radio-group">
          {{ $units := .FormValues.GetOrDefault "units" "imperial" }}
          {{ range .UnitOptions }}
            <label class="radio">
              <input type="radio" name="units" value="{{ .Value }}" {{ if eq $units .Value }}checked{{ end }}>
              <span>{{ .Label }}</span>
            </label>
          {{ end }}
        </div>
      </div>
      <div class="form-group">
        <label>Time format</label>
        <div class="radio-group">
          {{ $timefmt := .FormValues.GetOrDefault "time_format" "12h" }}
          {{ range .TimeFormatOptions }}
            <label class="radio">
              <input type="radio" name="time_format" value="{{ .Value }}" {{ if eq $timefmt .Value }}checked{{ end }} required>
              <span>{{ .Label }}</span>
            </label>
          {{ end }}
        </div>
      </div>
      <div class="form-group">
        <label>Show seconds</label>
        <div class="radio-group">
          {{ $showSeconds := .FormValues.GetOrDefault "show_seconds" "off" }}
          <label class="radio">
            <input type="radio" name="show_seconds" value="on" {{ if eq $showSeconds "on" }}checked{{ end }}>
            <span>Enabled</span>
          </label>
          <label class="radio">
            <input type="radio" name="show_seconds" value="off" {{ if eq $showSeconds "off" }}checked{{ end }}>
            <span>Disabled</span>
          </label>
        </div>
      </div>
      <div class="form-group">
        <label>Show location name</label>
        <div class="radio-group">
          {{ $showLocation := .FormValues.GetOrDefault "show_location_name" .DefaultShowLocation }}
          <label class="radio">
            <input type="radio" name="show_location_name" value="on" {{ if eq $showLocation "on" }}checked{{ end }}>
            <span>Enabled</span>
          </label>
          <label class="radio">
            <input type="radio" name="show_location_name" value="off" {{ if eq $showLocation "off" }}checked{{ end }}>
            <span>Disabled</span>
          </label>
        </div>
      </div>
      <div class="form-group">
        <label>Extra detail display</label>
        <div class="detail-grid">
          {{ range .DetailFields }}
            <div class="detail-setting">
              <span>{{ .Label }}</span>
              {{ $current := $.FormValues.GetOrDefault .Name (index $.DefaultDetailSettings .Name) }}
              <select name="{{ .Name }}">
                {{ range $.DetailOptions }}
                  <option value="{{ .Value }}" {{ if eq $current .Value }}selected{{ end }}>
                    {{ .Label }}
                  </option>
                {{ end }}
              </select>
            </div>
          {{ end }}
        </div>
      </div>
      <div class="form-group">
        <label for="theme">Theme</label>
        {{ $themeValue := .FormValues.GetOrDefault "theme" .DefaultTheme }}
        <select id="theme" name="theme">
          {{ range .ThemeOptions }}
            <option value="{{ .Value }}" {{ if eq $themeValue .Value }}selected{{ end }}>
              {{ .Label }}
            </option>
          {{ end }}
        </select>
      </div>
      <button type="submit">Save settings</button>
    </form>
  </div>
</body>
</html>`
