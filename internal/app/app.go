package app

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"nv-copa/internal/copa"
	"nv-copa/internal/db"

	"golang.org/x/crypto/bcrypt"
)

type App struct {
	store     *db.Store
	templates *template.Template
}

const activePredictionRound = 3

type PageData struct {
	User                 db.User
	Podium               db.PodiumPrediction
	Teams                []string
	Groups               []GroupView
	ThirdCount           int
	SelectedGroup        string
	CurrentRound         int
	RoundLabel           string
	GroupStandings       []StandingView
	CurrentRoundFixtures []FixtureView
	RoundPredicted       int
	RoundTotal           int
	Error                string
	Notice               string
}

type AuthData struct {
	Error  string
	Notice string
}

type PodiumData struct {
	User   db.User
	Teams  []string
	Error  string
	Notice string
}

type RankingPageData struct {
	User     db.User
	Rankings []db.UserRanking
}

type KnockoutPhase struct {
	Key      string
	Label    string
	Round    int
	Active   bool
	HasGames bool
	Waiting  string
}

type KnockoutPageData struct {
	User         db.User
	Fixtures     []FixtureView
	Predicted    int
	Total        int
	Phases       []KnockoutPhase
	CurrentPhase KnockoutPhase
	PrevPhase    string
	NextPhase    string
	Error        string
	Notice       string
}

type GroupPicksData struct {
	User       db.User
	Groups     []GroupView
	ThirdCount int
	Error      string
	Notice     string
}

type GroupsPageData struct {
	User                 db.User
	SelectedGroup        string
	SelectedRound        int
	RoundLabel           string
	CurrentRound         int
	CurrentRoundLabel    string
	Groups               []GroupView
	SelectedGroupTeams   []TeamView
	GroupStandings       []StandingView
	CurrentRoundFixtures []FixtureView
	ActiveRoundFixtures  []FixtureView
	ActiveRoundPredicted int
	ActiveRoundTotal     int
	ThirdCount           int
	Locked               bool
	GroupQualifiers      map[string]int
	Error                string
	Notice               string
}

type GroupView struct {
	Name  string
	Teams []TeamView
}

type TeamView struct {
	Name     string
	Display  string
	Flag     string
	Position int
}

type StandingView struct {
	Position     int
	TeamName     string
	TeamDisplay  string
	TeamFlag     string
	SortIndex    int
	Points       int
	Played       int
	Wins         int
	Draws        int
	Losses       int
	GoalsFor     int
	GoalsAgainst int
	GoalDiff     int
}

type FixtureView struct {
	ID            int64
	GroupName     string
	MatchDate     string
	MatchTime     string
	HomeTeam      string
	HomeDisplay   string
	HomeFlag      string
	AwayTeam      string
	AwayDisplay   string
	AwayFlag      string
	RealHomeScore string
	RealAwayScore string
	PredHomeScore string
	PredAwayScore string
	Locked            bool
	HasResult         bool
}

func New(store *db.Store) http.Handler {
	funcs := template.FuncMap{
		"dateTime": func(t time.Time) string {
			return t.Local().Format("02/01 15:04")
		},
		"datetimeInput": func(t time.Time) string {
			return t.Local().Format("2006-01-02T15:04")
		},
		"isClosed": func(t time.Time) bool {
			return time.Now().UTC().After(t.UTC())
		},
		"teamDisplay": copa.TeamDisplay,
		"qualifierPos": func(qualifiers map[string]int, teamName string) int {
			return qualifiers[teamName]
		},
	}

	templates := template.Must(template.New("").Funcs(funcs).ParseGlob("web/templates/*.html"))
	app := &App{store: store, templates: templates}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/login", app.login)
	mux.HandleFunc("/signup", app.signup)
	mux.HandleFunc("/logout", app.logout)
	mux.HandleFunc("/podium", app.savePodium)
	mux.HandleFunc("/group-picks", app.groupPicksPage)
	mux.HandleFunc("/group-picks/save", app.saveGroupPicks)
	mux.HandleFunc("/groups", app.groupsPage)
	mux.HandleFunc("/groups/results", app.saveFixtureResults)
	mux.HandleFunc("/groups/results/reset", app.resetFixtureResults)
	mux.HandleFunc("/groups/qualifiers", app.saveGroupQualifiers)
	mux.HandleFunc("/ranking", app.rankingPage)
	mux.HandleFunc("/knockout", app.knockoutPage)
	mux.HandleFunc("/knockout/predictions", app.saveKnockoutPredictions)
	mux.HandleFunc("/knockout/results", app.saveKnockoutResults)
	mux.HandleFunc("/round-predictions", app.saveRoundPredictions)
	mux.HandleFunc("/admin/predictions", app.adminPredictions)
	return mux
}

func (a *App) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	user, ok := a.requireAuth(w, r)
	if !ok {
		return
	}
	hasPodium, err := a.store.HasPodiumPrediction(user.ID)
	if err != nil {
		log.Printf("load podium status: %v", err)
		http.Error(w, "erro ao carregar palpite", http.StatusInternalServerError)
		return
	}
	if !hasPodium {
		a.render(w, "podium.html", PodiumData{User: user, Teams: copa.TeamNames()})
		return
	}
	lockedGroups, err := a.store.HasLockedGroupPredictions(user.ID)
	if err != nil {
		log.Printf("load group lock status: %v", err)
		http.Error(w, "erro ao carregar palpites de grupos", http.StatusInternalServerError)
		return
	}
	if !lockedGroups {
		http.Redirect(w, r, "/group-picks", http.StatusSeeOther)
		return
	}
	data := a.pageData(r, user)
	a.render(w, "layout.html", data)
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if _, ok := a.currentUser(r); ok {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		a.render(w, "auth.html", AuthData{})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	password := r.FormValue("password")
	user, passwordHash, err := a.store.UserPasswordHash(name)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) != nil {
		a.render(w, "auth.html", AuthData{Error: "nome ou senha invalidos"})
		return
	}

	if err := a.startSession(w, user.ID); err != nil {
		log.Printf("start session: %v", err)
		a.render(w, "auth.html", AuthData{Error: "nao foi possivel iniciar sessao"})
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) signup(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if _, ok := a.currentUser(r); ok {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		a.render(w, "signup.html", AuthData{})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	password := r.FormValue("password")
	if name == "" || password == "" {
		a.render(w, "signup.html", AuthData{Error: "nome e senha sao obrigatorios"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("hash password: %v", err)
		a.render(w, "signup.html", AuthData{Error: "nao foi possivel criar usuario"})
		return
	}

	user, err := a.store.CreateUser(name, string(hash))
	if err != nil {
		a.render(w, "signup.html", AuthData{Error: "usuario ja existe ou dados invalidos"})
		return
	}
	if err := a.startSession(w, user.ID); err != nil {
		log.Printf("start session: %v", err)
		a.render(w, "signup.html", AuthData{Error: "usuario criado, mas login falhou"})
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("nv_copa_session")
	if err == nil {
		_ = a.store.DeleteSession(cookie.Value)
	}
	http.SetCookie(w, expiredSessionCookie())
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (a *App) savePodium(w http.ResponseWriter, r *http.Request) {
	user, ok := a.requireAuth(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	champion := strings.TrimSpace(r.FormValue("champion"))
	runnerUp := strings.TrimSpace(r.FormValue("runner_up"))
	third := strings.TrimSpace(r.FormValue("third"))
	if !copa.ValidTeam(champion) || !copa.ValidTeam(runnerUp) || !copa.ValidTeam(third) {
		a.render(w, "podium.html", PodiumData{
			User:  user,
			Teams: copa.TeamNames(),
			Error: "selecao invalida no podio",
		})
		return
	}

	if err := a.store.SavePodiumPrediction(user.ID, champion, runnerUp, third); err != nil {
		a.render(w, "podium.html", PodiumData{
			User:  user,
			Teams: copa.TeamNames(),
			Error: err.Error(),
		})
		return
	}

	http.Redirect(w, r, "/group-picks", http.StatusSeeOther)
}

func (a *App) groupPicksPage(w http.ResponseWriter, r *http.Request) {
	user, ok := a.requireAuth(w, r)
	if !ok {
		return
	}
	hasPodium, err := a.store.HasPodiumPrediction(user.ID)
	if err != nil {
		log.Printf("load podium status for group picks: %v", err)
		http.Error(w, "erro ao carregar palpite", http.StatusInternalServerError)
		return
	}
	if !hasPodium {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	locked, err := a.store.HasLockedGroupPredictions(user.ID)
	if err != nil {
		log.Printf("load group lock for group picks: %v", err)
		http.Error(w, "erro ao carregar palpites de grupos", http.StatusInternalServerError)
		return
	}
	if locked {
		http.Redirect(w, r, "/groups", http.StatusSeeOther)
		return
	}
	if r.Method == http.MethodGet {
		a.render(w, "group_picks.html", a.groupPicksData(user))
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (a *App) saveGroupPicks(w http.ResponseWriter, r *http.Request) {
	user, ok := a.requireAuth(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	locked, err := a.store.HasLockedGroupPredictions(user.ID)
	if err != nil {
		http.Error(w, "erro ao carregar palpites de grupos", http.StatusInternalServerError)
		return
	}
	if locked {
		http.Redirect(w, r, "/groups", http.StatusSeeOther)
		return
	}
	predictions, err := parseGroupPredictions(r)
	if err != nil {
		data := a.groupPicksDataFromRequest(user, r)
		data.Error = err.Error()
		a.render(w, "group_picks.html", data)
		return
	}
	if err := a.store.SaveGroupPredictions(user.ID, predictions); err != nil {
		data := a.groupPicksDataFromRequest(user, r)
		data.Error = err.Error()
		a.render(w, "group_picks.html", data)
		return
	}
	if err := a.store.FinalizeGroupPredictions(user.ID); err != nil {
		data := a.groupPicksDataFromRequest(user, r)
		data.Error = err.Error()
		a.render(w, "group_picks.html", data)
		return
	}
	http.Redirect(w, r, "/groups", http.StatusSeeOther)
}

func (a *App) groupsPage(w http.ResponseWriter, r *http.Request) {
	user, ok := a.requireAuth(w, r)
	if !ok {
		return
	}
	hasPodium, err := a.store.HasPodiumPrediction(user.ID)
	if err != nil {
		log.Printf("load podium status in groups page: %v", err)
		http.Error(w, "erro ao carregar palpite", http.StatusInternalServerError)
		return
	}
	if !hasPodium {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	locked, err := a.store.HasLockedGroupPredictions(user.ID)
	if err != nil {
		log.Printf("load group lock in groups page: %v", err)
		http.Error(w, "erro ao carregar palpites de grupos", http.StatusInternalServerError)
		return
	}
	if !locked {
		http.Redirect(w, r, "/group-picks", http.StatusSeeOther)
		return
	}
	if r.Method == http.MethodGet {
		data := a.groupsPageData(r, user)
		a.render(w, "groups.html", data)
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

var knockoutPhases = []KnockoutPhase{
	{Key: "16avos", Label: "16 avos de final", Round: 4},
	{Key: "oitavas", Label: "Oitavas de final", Round: 5, Waiting: "Aguardando resultados das 16 avos de final."},
	{Key: "quartas", Label: "Quartas de final", Round: 6, Waiting: "Aguardando resultados das oitavas de final."},
	{Key: "semis", Label: "Semifinais", Round: 7, Waiting: "Aguardando resultados das quartas de final."},
	{Key: "terceiro", Label: "Disputa de 3o lugar", Round: 8, Waiting: "Aguardando resultados das semifinais."},
	{Key: "final", Label: "Final", Round: 9, Waiting: "Aguardando resultados das semifinais."},
}

func (a *App) knockoutPage(w http.ResponseWriter, r *http.Request) {
	user, ok := a.requireAuth(w, r)
	if !ok {
		return
	}
	phase := r.URL.Query().Get("phase")
	data := a.knockoutDataForPhase(user, phase)
	a.render(w, "knockout.html", data)
}

func (a *App) saveKnockoutPredictions(w http.ResponseWriter, r *http.Request) {
	user, ok := a.requireAuth(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	phase := r.FormValue("phase")
	if phase == "" {
		phase = "16avos"
	}
	predictions, err := parseRoundPredictions(r)
	if err != nil {
		data := a.knockoutDataForPhase(user, phase)
		data.Error = err.Error()
		a.render(w, "knockout.html", data)
		return
	}
	if len(predictions) == 0 {
		data := a.knockoutDataForPhase(user, phase)
		data.Error = "preencha ao menos um palpite completo para salvar"
		a.render(w, "knockout.html", data)
		return
	}
	if err := a.store.SaveFixturePredictions(user.ID, predictions, time.Now()); err != nil {
		data := a.knockoutDataForPhase(user, phase)
		data.Error = err.Error()
		a.render(w, "knockout.html", data)
		return
	}
	data := a.knockoutDataForPhase(user, phase)
	data.Notice = "Palpites salvos."
	a.render(w, "knockout.html", data)
}

func (a *App) knockoutDataForPhase(user db.User, phaseKey string) KnockoutPageData {
	idx := 0
	for i, p := range knockoutPhases {
		if p.Key == phaseKey {
			idx = i
			break
		}
	}
	current := knockoutPhases[idx]

	fixtures, _ := a.store.KnockoutFixturePredictions(user.ID, current.Round)
	predicted, total, _ := a.store.KnockoutPredictionProgress(user.ID, current.Round)

	current.Active = true
	current.HasGames = len(fixtures) > 0

	phases := make([]KnockoutPhase, len(knockoutPhases))
	copy(phases, knockoutPhases)
	phases[idx].Active = true

	var prev, next string
	if idx > 0 {
		prev = knockoutPhases[idx-1].Key
	}
	if idx < len(knockoutPhases)-1 {
		next = knockoutPhases[idx+1].Key
	}

	return KnockoutPageData{
		User:         user,
		Fixtures:     fixtureViews(fixtures),
		Predicted:    predicted,
		Total:        total,
		Phases:       phases,
		CurrentPhase: current,
		PrevPhase:    prev,
		NextPhase:    next,
	}
}

func (a *App) saveKnockoutResults(w http.ResponseWriter, r *http.Request) {
	user, ok := a.requireAuth(w, r)
	if !ok {
		return
	}
	if !user.IsAdmin {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	phase := r.FormValue("phase")
	if phase == "" {
		phase = "16avos"
	}
	results, err := parseFixtureResults(r)
	if err != nil {
		data := a.knockoutDataForPhase(user, phase)
		data.Error = err.Error()
		a.render(w, "knockout.html", data)
		return
	}
	for fixtureID, score := range results {
		if err := a.store.SetFixtureResult(fixtureID, score[0], score[1]); err != nil {
			data := a.knockoutDataForPhase(user, phase)
			data.Error = err.Error()
			a.render(w, "knockout.html", data)
			return
		}
	}
	data := a.knockoutDataForPhase(user, phase)
	data.Notice = "Resultados oficiais atualizados."
	a.render(w, "knockout.html", data)
}

func (a *App) saveRoundPredictions(w http.ResponseWriter, r *http.Request) {
	user, ok := a.requireAuth(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	selectedGroup := normalizeGroupName(r.FormValue("selected_group"))
	if selectedGroup == "" {
		selectedGroup = "Grupo A"
	}
	selectedRound, _ := strconv.Atoi(r.FormValue("selected_round"))
	if selectedRound < 1 || selectedRound > 3 {
		selectedRound = 1
	}
	renderGroupsPage := r.FormValue("return_to") == "groups"
	predictions, err := parseRoundPredictions(r)
	if err != nil {
		if renderGroupsPage {
			data := a.groupsPageDataFor(user, selectedGroup, selectedRound)
			data.Error = err.Error()
			a.render(w, "groups.html", data)
			return
		}
		data := a.loadData(user, selectedGroup)
		data.Error = err.Error()
		a.renderMain(w, data)
		return
	}
	if len(predictions) == 0 {
		if renderGroupsPage {
			data := a.groupsPageDataFor(user, selectedGroup, selectedRound)
			data.Error = "preencha ao menos um palpite completo para salvar"
			a.render(w, "groups.html", data)
			return
		}
		data := a.loadData(user, selectedGroup)
		data.Error = "preencha ao menos um palpite completo para salvar"
		a.renderMain(w, data)
		return
	}

	if err := a.store.SaveFixturePredictions(user.ID, predictions, time.Now()); err != nil {
		if renderGroupsPage {
			data := a.groupsPageDataFor(user, selectedGroup, selectedRound)
			data.Error = err.Error()
			a.render(w, "groups.html", data)
			return
		}
		data := a.loadData(user, selectedGroup)
		data.Error = err.Error()
		a.renderMain(w, data)
		return
	}

	if renderGroupsPage {
		data := a.groupsPageDataFor(user, selectedGroup, selectedRound)
		data.Notice = "Palpites da rodada salvos."
		a.render(w, "groups.html", data)
		return
	}
	data := a.loadData(user, selectedGroup)
	data.Notice = "Palpites da rodada salvos."
	a.renderMain(w, data)
}

func (a *App) pageData(r *http.Request, user db.User) PageData {
	selectedGroup := normalizeGroupName(r.URL.Query().Get("group"))
	if selectedGroup == "" {
		selectedGroup = "Grupo A"
	}
	return a.loadData(user, selectedGroup)
}

func (a *App) loadData(user db.User, selectedGroup string) PageData {
	podium, err := a.store.PodiumPrediction(user.ID)
	if err != nil {
		log.Printf("load podium: %v", err)
	}

	groups, thirdCount := a.groupViews(user.ID)
	currentRound := activePredictionRound
	fixtures, err := a.store.GroupFixturePredictions(user.ID, currentRound, selectedGroup)
	if err != nil {
		log.Printf("load fixtures: %v", err)
	}
	roundPredicted, roundTotal, err := a.store.RoundPredictionProgress(user.ID, currentRound)
	if err != nil {
		log.Printf("load round progress: %v", err)
	}
	data := PageData{
		User:                 user,
		Podium:               podium,
		Teams:                copa.TeamNames(),
		Groups:               groups,
		ThirdCount:           thirdCount,
		SelectedGroup:        selectedGroup,
		CurrentRound:         currentRound,
		RoundLabel:           strconv.Itoa(currentRound) + "a Rodada",
		GroupStandings:       buildStandings(a.store, selectedGroup),
		CurrentRoundFixtures: fixtureViews(fixtures),
		RoundPredicted:       roundPredicted,
		RoundTotal:           roundTotal,
	}
	return data
}

func (a *App) groupsPageData(r *http.Request, user db.User) GroupsPageData {
	selectedGroup := normalizeGroupName(r.URL.Query().Get("group"))
	if selectedGroup == "" {
		selectedGroup = "Grupo A"
	}
	selectedRound, _ := strconv.Atoi(r.URL.Query().Get("round"))
	if selectedRound < 1 || selectedRound > 3 {
		selectedRound = activePredictionRound
	}
	return a.groupsPageDataFor(user, selectedGroup, selectedRound)
}

func (a *App) groupPicksData(user db.User) GroupPicksData {
	groups, thirdCount := a.groupViews(user.ID)
	return GroupPicksData{
		User:       user,
		Groups:     groups,
		ThirdCount: thirdCount,
	}
}

func (a *App) groupPicksDataFromRequest(user db.User, r *http.Request) GroupPicksData {
	data := a.groupPicksData(user)
	selected, thirdCount := parseLooseGroupSelections(r)
	if len(selected) == 0 {
		return data
	}
	for groupIdx := range data.Groups {
		for teamIdx := range data.Groups[groupIdx].Teams {
			key := data.Groups[groupIdx].Name + "\x00" + data.Groups[groupIdx].Teams[teamIdx].Name
			if position, ok := selected[key]; ok {
				data.Groups[groupIdx].Teams[teamIdx].Position = position
			}
		}
	}
	data.ThirdCount = thirdCount
	return data
}

func (a *App) groupsPageDataFor(user db.User, selectedGroup string, selectedRound int) GroupsPageData {
	groups, thirdCount := a.groupViews(user.ID)
	locked, err := a.store.HasLockedGroupPredictions(user.ID)
	if err != nil {
		log.Printf("load group lock: %v", err)
	}
	fixtures, err := a.store.GroupFixturePredictions(user.ID, selectedRound, selectedGroup)
	if err != nil {
		log.Printf("load group page fixtures: %v", err)
	}
	currentRound := activePredictionRound
	activeFixtures, err := a.store.GroupFixturePredictions(user.ID, currentRound, selectedGroup)
	if err != nil {
		log.Printf("load active round fixtures in groups page: %v", err)
	}
	activePredicted, activeTotal := fixturePredictionProgress(activeFixtures)
	groupResults, err := a.store.GroupResults()
	if err != nil {
		log.Printf("load group results: %v", err)
	}
	qualifiers := make(map[string]int)
	for _, r := range groupResults {
		if r.GroupName == selectedGroup {
			qualifiers[r.TeamName] = r.Position
		}
	}
	return GroupsPageData{
		User:                 user,
		SelectedGroup:        selectedGroup,
		SelectedRound:        selectedRound,
		RoundLabel:           strconv.Itoa(selectedRound) + "a Rodada",
		CurrentRound:         currentRound,
		CurrentRoundLabel:    strconv.Itoa(currentRound) + "a Rodada",
		Groups:               groups,
		SelectedGroupTeams:   selectedGroupTeams(groups, selectedGroup),
		GroupStandings:       buildStandings(a.store, selectedGroup),
		CurrentRoundFixtures: fixtureViews(fixtures),
		ActiveRoundFixtures:  fixtureViews(activeFixtures),
		ActiveRoundPredicted: activePredicted,
		ActiveRoundTotal:     activeTotal,
		ThirdCount:           thirdCount,
		Locked:               locked,
		GroupQualifiers:      qualifiers,
	}
}

func (a *App) saveSelectedGroup(w http.ResponseWriter, r *http.Request) {
	user, ok := a.requireAuth(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	locked, err := a.store.HasLockedGroupPredictions(user.ID)
	if err != nil {
		http.Error(w, "erro ao carregar palpites de grupos", http.StatusInternalServerError)
		return
	}
	selectedGroup := normalizeGroupName(r.FormValue("selected_group"))
	if selectedGroup == "" {
		selectedGroup = "Grupo A"
	}
	selectedRound, _ := strconv.Atoi(r.FormValue("selected_round"))
	if selectedRound < 1 || selectedRound > 3 {
		selectedRound = 1
	}
	if locked {
		data := a.groupsPageDataFor(user, selectedGroup, selectedRound)
		data.Error = "o palpite dos grupos ja foi finalizado e nao pode mais ser alterado"
		a.render(w, "groups.html", data)
		return
	}
	predictions, err := parseSingleGroupPrediction(r, selectedGroup)
	if err != nil {
		data := a.groupsPageDataFor(user, selectedGroup, selectedRound)
		data.Error = err.Error()
		a.render(w, "groups.html", data)
		return
	}
	if err := a.store.SaveGroupPredictionsForGroup(user.ID, selectedGroup, predictions); err != nil {
		data := a.groupsPageDataFor(user, selectedGroup, selectedRound)
		data.Error = err.Error()
		a.render(w, "groups.html", data)
		return
	}
	data := a.groupsPageDataFor(user, selectedGroup, selectedRound)
	data.Notice = "Palpite do grupo salvo."
	a.render(w, "groups.html", data)
}

func (a *App) finalizeGroups(w http.ResponseWriter, r *http.Request) {
	user, ok := a.requireAuth(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	selectedGroup := normalizeGroupName(r.FormValue("selected_group"))
	if selectedGroup == "" {
		selectedGroup = "Grupo A"
	}
	selectedRound, _ := strconv.Atoi(r.FormValue("selected_round"))
	if selectedRound < 1 || selectedRound > 3 {
		selectedRound = 1
	}
	if err := a.store.FinalizeGroupPredictions(user.ID); err != nil {
		data := a.groupsPageDataFor(user, selectedGroup, selectedRound)
		data.Error = err.Error()
		a.render(w, "groups.html", data)
		return
	}
	data := a.groupsPageDataFor(user, selectedGroup, selectedRound)
	data.Notice = "Palpite dos grupos finalizado e bloqueado."
	a.render(w, "groups.html", data)
}

func (a *App) saveFixtureResults(w http.ResponseWriter, r *http.Request) {
	user, ok := a.requireAuth(w, r)
	if !ok {
		return
	}
	if !user.IsAdmin {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	selectedGroup := normalizeGroupName(r.FormValue("selected_group"))
	if selectedGroup == "" {
		selectedGroup = "Grupo A"
	}
	selectedRound, _ := strconv.Atoi(r.FormValue("selected_round"))
	if selectedRound < 1 || selectedRound > 3 {
		selectedRound = 1
	}
	results, err := parseFixtureResults(r)
	if err != nil {
		data := a.groupsPageDataFor(user, selectedGroup, selectedRound)
		data.Error = err.Error()
		a.render(w, "groups.html", data)
		return
	}
	for fixtureID, score := range results {
		if err := a.store.SetFixtureResult(fixtureID, score[0], score[1]); err != nil {
			data := a.groupsPageDataFor(user, selectedGroup, selectedRound)
			data.Error = err.Error()
			a.render(w, "groups.html", data)
			return
		}
	}
	data := a.groupsPageDataFor(user, selectedGroup, selectedRound)
	data.Notice = "Resultados oficiais atualizados."
	a.render(w, "groups.html", data)
}

func (a *App) resetFixtureResults(w http.ResponseWriter, r *http.Request) {
	user, ok := a.requireAuth(w, r)
	if !ok {
		return
	}
	if !user.IsAdmin {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	selectedGroup := normalizeGroupName(r.FormValue("selected_group"))
	selectedRound, err := strconv.Atoi(r.FormValue("selected_round"))
	if selectedGroup == "" || selectedRound < 1 || selectedRound > 3 || err != nil {
		http.Error(w, "grupo ou rodada invalida", http.StatusBadRequest)
		return
	}
	fixtureID, err := strconv.ParseInt(r.FormValue("fixture_id"), 10, 64)
	if err != nil || fixtureID < 1 {
		http.Error(w, "partida invalida", http.StatusBadRequest)
		return
	}
	if err := a.store.ResetFixtureResult(fixtureID, selectedGroup, selectedRound); err != nil {
		log.Printf("reset fixture results: %v", err)
		http.Error(w, "erro ao resetar resultados oficiais", http.StatusInternalServerError)
		return
	}

	data := a.groupsPageDataFor(user, selectedGroup, selectedRound)
	data.Notice = "Resultado oficial resetado. Os pontos correspondentes foram removidos do ranking."
	a.render(w, "groups.html", data)
}

func (a *App) rankingPage(w http.ResponseWriter, r *http.Request) {
	user, ok := a.requireAuth(w, r)
	if !ok {
		return
	}
	rankings, err := a.store.FullRanking()
	if err != nil {
		log.Printf("load ranking: %v", err)
		http.Error(w, "erro ao carregar ranking", http.StatusInternalServerError)
		return
	}
	a.render(w, "ranking.html", RankingPageData{User: user, Rankings: rankings})
}

func (a *App) saveGroupQualifiers(w http.ResponseWriter, r *http.Request) {
	user, ok := a.requireAuth(w, r)
	if !ok {
		return
	}
	if !user.IsAdmin {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	selectedGroup := normalizeGroupName(r.FormValue("selected_group"))
	if selectedGroup == "" {
		http.Error(w, "grupo invalido", http.StatusBadRequest)
		return
	}
	selectedRound, _ := strconv.Atoi(r.FormValue("selected_round"))
	if selectedRound < 1 || selectedRound > 3 {
		selectedRound = 1
	}

	var results []db.GroupResult
	for _, group := range copa.Groups {
		if group.Name != selectedGroup {
			continue
		}
		for _, team := range group.Teams {
			field := "qualifier_" + team.Name
			value := strings.TrimSpace(r.FormValue(field))
			if value == "" {
				continue
			}
			position, err := strconv.Atoi(value)
			if err != nil || position < 1 || position > 3 {
				data := a.groupsPageDataFor(user, selectedGroup, selectedRound)
				data.Error = "posicao invalida para " + team.Name
				a.render(w, "groups.html", data)
				return
			}
			results = append(results, db.GroupResult{
				GroupName: selectedGroup,
				TeamName:  team.Name,
				Position:  position,
			})
		}
	}

	if err := a.store.SaveGroupResults(selectedGroup, results); err != nil {
		data := a.groupsPageDataFor(user, selectedGroup, selectedRound)
		data.Error = err.Error()
		a.render(w, "groups.html", data)
		return
	}

	data := a.groupsPageDataFor(user, selectedGroup, selectedRound)
	data.Notice = "Classificados do grupo atualizados."
	a.render(w, "groups.html", data)
}

func (a *App) groupViews(userID int64) ([]GroupView, int) {
	predictions, err := a.store.GroupPredictions(userID)
	if err != nil {
		log.Printf("load group predictions: %v", err)
	}
	selected := make(map[string]int)
	thirdCount := 0
	for _, prediction := range predictions {
		key := prediction.GroupName + "\x00" + prediction.TeamName
		selected[key] = prediction.Position
		if prediction.Position == 3 {
			thirdCount++
		}
	}

	var groups []GroupView
	for _, group := range copa.Groups {
		view := GroupView{Name: group.Name}
		for _, team := range group.Teams {
			key := group.Name + "\x00" + team.Name
			view.Teams = append(view.Teams, TeamView{
				Name:     team.Name,
				Display:  copa.TeamDisplay(team.Name),
				Flag:     team.Flag,
				Position: selected[key],
			})
		}
		groups = append(groups, view)
	}
	return groups, thirdCount
}

func buildStandings(store *db.Store, groupName string) []StandingView {
	fixtures, err := store.AllGroupFixtures(groupName)
	if err != nil {
		log.Printf("load standings fixtures: %v", err)
	}

	rows := make(map[string]*StandingView)
	for _, group := range copa.Groups {
		if group.Name != groupName {
			continue
		}
		for idx, team := range group.Teams {
			rows[team.Name] = &StandingView{
				TeamName:    team.Name,
				TeamDisplay: copa.TeamDisplay(team.Name),
				TeamFlag:    team.Flag,
				SortIndex:   idx,
			}
		}
	}

	for _, fixture := range fixtures {
		home := rows[fixture.HomeTeam]
		away := rows[fixture.AwayTeam]
		if home == nil || away == nil {
			continue
		}
		if !fixture.HomeScore.Valid || !fixture.AwayScore.Valid {
			continue
		}
		homeGoals := int(fixture.HomeScore.Int64)
		awayGoals := int(fixture.AwayScore.Int64)

		home.Played++
		away.Played++
		home.GoalsFor += homeGoals
		home.GoalsAgainst += awayGoals
		away.GoalsFor += awayGoals
		away.GoalsAgainst += homeGoals
		home.GoalDiff = home.GoalsFor - home.GoalsAgainst
		away.GoalDiff = away.GoalsFor - away.GoalsAgainst

		switch {
		case homeGoals > awayGoals:
			home.Wins++
			home.Points += 3
			away.Losses++
		case homeGoals < awayGoals:
			away.Wins++
			away.Points += 3
			home.Losses++
		default:
			home.Draws++
			away.Draws++
			home.Points++
			away.Points++
		}
	}

	var standings []StandingView
	for _, row := range rows {
		standings = append(standings, *row)
	}
	sortStandings(standings)
	for idx := range standings {
		standings[idx].Position = idx + 1
	}
	return standings
}

func sortStandings(standings []StandingView) {
	for i := 0; i < len(standings); i++ {
		for j := i + 1; j < len(standings); j++ {
			swap := false
			switch {
			case standings[j].Points > standings[i].Points:
				swap = true
			case standings[j].Points == standings[i].Points && standings[j].GoalDiff > standings[i].GoalDiff:
				swap = true
			case standings[j].Points == standings[i].Points && standings[j].GoalDiff == standings[i].GoalDiff && standings[j].GoalsFor > standings[i].GoalsFor:
				swap = true
			case standings[j].Points == standings[i].Points && standings[j].GoalDiff == standings[i].GoalDiff && standings[j].GoalsFor == standings[i].GoalsFor && standings[j].SortIndex < standings[i].SortIndex:
				swap = true
			}
			if swap {
				standings[i], standings[j] = standings[j], standings[i]
			}
		}
	}
}

func fixtureViews(fixtures []db.FixturePrediction) []FixtureView {
	var views []FixtureView
	now := time.Now()
	for _, fixture := range fixtures {
		homeTeam, _ := copa.TeamByName(fixture.HomeTeam)
		awayTeam, _ := copa.TeamByName(fixture.AwayTeam)
		matchTime := ""
		if fixture.Round >= 4 {
			matchTime = fixture.MatchDate.Format("15:04")
		}
		view := FixtureView{
			ID:          fixture.ID,
			GroupName:   fixture.GroupName,
			MatchDate:   fixture.MatchDate.Format("02/01"),
			MatchTime:   matchTime,
			HomeTeam:    fixture.HomeTeam,
			HomeDisplay: copa.TeamDisplay(fixture.HomeTeam),
			HomeFlag:    homeTeam.Flag,
			AwayTeam:    fixture.AwayTeam,
			AwayDisplay: copa.TeamDisplay(fixture.AwayTeam),
			AwayFlag:    awayTeam.Flag,
			Locked:      !isPredictionOpen(now, fixture.MatchDate),
			HasResult:   fixture.HomeScore.Valid && fixture.AwayScore.Valid,
		}
		if fixture.PredHomeScore.Valid {
			view.PredHomeScore = strconv.FormatInt(fixture.PredHomeScore.Int64, 10)
		}
		if fixture.PredAwayScore.Valid {
			view.PredAwayScore = strconv.FormatInt(fixture.PredAwayScore.Int64, 10)
		}
		if fixture.HomeScore.Valid {
			view.RealHomeScore = strconv.FormatInt(fixture.HomeScore.Int64, 10)
		}
		if fixture.AwayScore.Valid {
			view.RealAwayScore = strconv.FormatInt(fixture.AwayScore.Int64, 10)
		}
		views = append(views, view)
	}
	return views
}

func fixturePredictionProgress(fixtures []db.FixturePrediction) (int, int) {
	predicted := 0
	for _, fixture := range fixtures {
		if fixture.PredHomeScore.Valid && fixture.PredAwayScore.Valid {
			predicted++
		}
	}
	return predicted, len(fixtures)
}

func (a *App) renderMain(w http.ResponseWriter, data PageData) {
	if err := a.templates.ExecuteTemplate(w, "main", data); err != nil {
		log.Printf("render main: %v", err)
	}
}

func (a *App) render(w http.ResponseWriter, name string, data any) {
	if err := a.templates.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("render %s: %v", name, err)
	}
}

func parseGroupPredictions(r *http.Request) ([]db.GroupPrediction, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	byGroupPosition := make(map[string]bool)
	thirdCount := 0
	var predictions []db.GroupPrediction
	for _, group := range copa.Groups {
		for _, team := range group.Teams {
			field := "pick_" + group.Name + "_" + team.Name
			value := r.FormValue(field)
			if value == "" {
				continue
			}

			position, err := strconv.Atoi(value)
			if err != nil || position < 1 || position > 3 {
				return nil, errInvalidGroupPick()
			}
			if !copa.ValidGroupTeam(group.Name, team.Name) {
				return nil, errInvalidGroupPick()
			}

			slotKey := group.Name + "_" + value
			if byGroupPosition[slotKey] {
				return nil, errDuplicateGroupPosition(group.Name)
			}
			byGroupPosition[slotKey] = true
			if position == 3 {
				thirdCount++
			}

			predictions = append(predictions, db.GroupPrediction{
				UserID:    0,
				GroupName: group.Name,
				TeamName:  team.Name,
				Position:  position,
			})
		}
	}

	for _, group := range copa.Groups {
		if !byGroupPosition[group.Name+"_1"] || !byGroupPosition[group.Name+"_2"] {
			return nil, errMissingGroupLeaders(group.Name)
		}
	}
	if thirdCount != 8 {
		return nil, errThirdCount(thirdCount)
	}

	return predictions, nil
}

func parseLooseGroupSelections(r *http.Request) (map[string]int, int) {
	if err := r.ParseForm(); err != nil {
		return nil, 0
	}

	selected := make(map[string]int)
	thirdCount := 0
	for _, group := range copa.Groups {
		for _, team := range group.Teams {
			field := "pick_" + group.Name + "_" + team.Name
			value := strings.TrimSpace(r.FormValue(field))
			if value == "" {
				continue
			}
			position, err := strconv.Atoi(value)
			if err != nil || position < 1 || position > 3 {
				continue
			}
			key := group.Name + "\x00" + team.Name
			selected[key] = position
			if position == 3 {
				thirdCount++
			}
		}
	}
	return selected, thirdCount
}

func parseSingleGroupPrediction(r *http.Request, groupName string) ([]db.GroupPrediction, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	positions := make(map[int]bool)
	var predictions []db.GroupPrediction
	for _, group := range copa.Groups {
		if group.Name != groupName {
			continue
		}
		for _, team := range group.Teams {
			field := "pick_" + group.Name + "_" + team.Name
			value := r.FormValue(field)
			if value == "" {
				continue
			}
			position, err := strconv.Atoi(value)
			if err != nil || position < 1 || position > 3 {
				return nil, errInvalidGroupPick()
			}
			if positions[position] {
				return nil, errDuplicateGroupPosition(groupName)
			}
			positions[position] = true
			predictions = append(predictions, db.GroupPrediction{
				GroupName: group.Name,
				TeamName:  team.Name,
				Position:  position,
			})
		}
	}
	if !positions[1] || !positions[2] {
		return nil, errMissingGroupLeaders(groupName)
	}
	return predictions, nil
}

func parseFixtureResults(r *http.Request) (map[int64][2]int64, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	results := make(map[int64][2]int64)
	for key := range r.Form {
		if !strings.HasPrefix(key, "result_") || !strings.HasSuffix(key, "_home") {
			continue
		}
		idText := strings.TrimSuffix(strings.TrimPrefix(key, "result_"), "_home")
		fixtureID, err := strconv.ParseInt(idText, 10, 64)
		if err != nil {
			return nil, validationError("partida invalida")
		}
		homeText := strings.TrimSpace(r.FormValue("result_" + idText + "_home"))
		awayText := strings.TrimSpace(r.FormValue("result_" + idText + "_away"))
		if homeText == "" && awayText == "" {
			continue
		}
		if homeText == "" || awayText == "" {
			return nil, validationError("preencha os dois placares do resultado oficial")
		}
		homeScore, err := strconv.ParseInt(homeText, 10, 64)
		if err != nil || homeScore < 0 {
			return nil, validationError("resultado oficial invalido")
		}
		awayScore, err := strconv.ParseInt(awayText, 10, 64)
		if err != nil || awayScore < 0 {
			return nil, validationError("resultado oficial invalido")
		}
		results[fixtureID] = [2]int64{homeScore, awayScore}
	}
	return results, nil
}

type validationError string

func (e validationError) Error() string {
	return string(e)
}

func errInvalidGroupPick() error {
	return validationError("palpite de grupo invalido")
}

func errDuplicateGroupPosition(groupName string) error {
	return validationError("cada posicao pode ser usada uma vez em " + groupName)
}

func errMissingGroupLeaders(groupName string) error {
	return validationError("escolha primeiro e segundo colocados em " + groupName)
}

func errThirdCount(count int) error {
	return validationError("escolha exatamente 8 terceiros classificados; agora voce escolheu " + strconv.Itoa(count))
}

func parseRoundPredictions(r *http.Request) (map[int64][2]int64, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	predictions := make(map[int64][2]int64)
	for key := range r.Form {
		if !strings.HasPrefix(key, "fixture_") || !strings.HasSuffix(key, "_home") {
			continue
		}
		idText := strings.TrimSuffix(strings.TrimPrefix(key, "fixture_"), "_home")
		fixtureID, err := strconv.ParseInt(idText, 10, 64)
		if err != nil {
			return nil, validationError("partida invalida")
		}
		homeText := strings.TrimSpace(r.FormValue("fixture_" + idText + "_home"))
		awayText := strings.TrimSpace(r.FormValue("fixture_" + idText + "_away"))
		if homeText == "" && awayText == "" {
			continue
		}
		if homeText == "" || awayText == "" {
			return nil, validationError("preencha os dois placares do jogo para salvar")
		}
		homeScore, err := strconv.ParseInt(homeText, 10, 64)
		if err != nil || homeScore < 0 {
			return nil, validationError("placar invalido")
		}
		awayScore, err := strconv.ParseInt(awayText, 10, 64)
		if err != nil || awayScore < 0 {
			return nil, validationError("placar invalido")
		}
		predictions[fixtureID] = [2]int64{homeScore, awayScore}
	}
	return predictions, nil
}

func normalizeGroupName(value string) string {
	for _, group := range copa.Groups {
		if group.Name == value {
			return value
		}
	}
	return ""
}

type AdminPredictionsData struct {
	User         db.User
	Users        []db.User
	SelectedUser *db.User
	Fixtures     []FixtureView
	Error        string
	Notice       string
}

func (a *App) adminPredictions(w http.ResponseWriter, r *http.Request) {
	user, ok := a.requireAuth(w, r)
	if !ok {
		return
	}
	if !user.IsAdmin {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	users, err := a.store.AllUsers()
	if err != nil {
		http.Error(w, "erro ao carregar usuarios", http.StatusInternalServerError)
		return
	}

	var filteredUsers []db.User
	for _, u := range users {
		if u.ID != user.ID {
			filteredUsers = append(filteredUsers, u)
		}
	}

	data := AdminPredictionsData{User: user, Users: filteredUsers}

	targetUserIDStr := r.FormValue("user_id")
	if targetUserIDStr == "" && r.URL.Query().Get("user_id") != "" {
		targetUserIDStr = r.URL.Query().Get("user_id")
	}

	if targetUserIDStr != "" {
		targetUserID, _ := strconv.ParseInt(targetUserIDStr, 10, 64)
		for _, u := range users {
			if u.ID == targetUserID {
				selected := u
				data.SelectedUser = &selected
				break
			}
		}

		if data.SelectedUser != nil && r.Method == http.MethodPost {
			predictions := make(map[int64][2]int64)
			for key, values := range r.Form {
				if !strings.HasPrefix(key, "fixture_") || !strings.HasSuffix(key, "_home") {
					continue
				}
				parts := strings.Split(key, "_")
				if len(parts) != 3 {
					continue
				}
				fixtureID, err := strconv.ParseInt(parts[1], 10, 64)
				if err != nil {
					continue
				}
				homeVal := strings.TrimSpace(values[0])
				awayVal := strings.TrimSpace(r.FormValue("fixture_" + parts[1] + "_away"))
				if homeVal == "" || awayVal == "" {
					continue
				}
				home, err := strconv.ParseInt(homeVal, 10, 64)
				if err != nil {
					continue
				}
				away, err := strconv.ParseInt(awayVal, 10, 64)
				if err != nil {
					continue
				}
				predictions[fixtureID] = [2]int64{home, away}
			}
			if len(predictions) > 0 {
				if err := a.store.AdminSaveFixturePredictions(targetUserID, predictions); err != nil {
					data.Error = err.Error()
				} else {
					data.Notice = "Palpites salvos para " + data.SelectedUser.Name + "."
				}
			}
		}

		if data.SelectedUser != nil {
			missing, err := a.store.MissingPredictions(data.SelectedUser.ID)
			if err != nil {
				http.Error(w, "erro ao carregar jogos pendentes", http.StatusInternalServerError)
				return
			}
			data.Fixtures = fixtureViews(missing)
		}
	}

	a.render(w, "admin_predictions.html", data)
}

func isPredictionOpen(now time.Time, matchDate time.Time) bool {
	return now.Before(matchDate)
}

func selectedGroupTeams(groups []GroupView, selectedGroup string) []TeamView {
	for _, group := range groups {
		if group.Name == selectedGroup {
			return group.Teams
		}
	}
	return nil
}

func (a *App) requireAuth(w http.ResponseWriter, r *http.Request) (db.User, bool) {
	user, ok := a.currentUser(r)
	if ok {
		return user, true
	}
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/login")
		w.WriteHeader(http.StatusUnauthorized)
		return db.User{}, false
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
	return db.User{}, false
}

func (a *App) currentUser(r *http.Request) (db.User, bool) {
	cookie, err := r.Cookie("nv_copa_session")
	if err != nil || cookie.Value == "" {
		return db.User{}, false
	}
	user, err := a.store.UserBySessionToken(cookie.Value)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("load session: %v", err)
		}
		return db.User{}, false
	}
	return user, true
}

func (a *App) startSession(w http.ResponseWriter, userID int64) error {
	token, err := randomToken()
	if err != nil {
		return err
	}
	expiresAt := time.Now().UTC().Add(30 * 24 * time.Hour)
	if err := a.store.CreateSession(userID, token, expiresAt); err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "nv_copa_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
	})
	return nil
}

func randomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func expiredSessionCookie() *http.Cookie {
	return &http.Cookie{
		Name:     "nv_copa_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	}
}
