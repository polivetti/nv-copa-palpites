package copa

type Team struct {
	Name    string
	Display string
	Flag    string
}

type Group struct {
	Name  string
	Teams []Team
}

var Groups = []Group{
	{Name: "Grupo A", Teams: []Team{
		{Name: "Mexico", Display: "Mexico", Flag: "🇲🇽"},
		{Name: "Africa do Sul", Display: "Africa do Sul", Flag: "🇿🇦"},
		{Name: "Coreia do Sul", Display: "Coreia do Sul", Flag: "🇰🇷"},
		{Name: "Republica Tcheca", Display: "Republica Tcheca", Flag: "🇨🇿"},
	}},
	{Name: "Grupo B", Teams: []Team{
		{Name: "Canada", Display: "Canada", Flag: "🇨🇦"},
		{Name: "Bosnia", Display: "Bosnia e Herzegovina", Flag: "🇧🇦"},
		{Name: "Catar", Display: "Catar", Flag: "🇶🇦"},
		{Name: "Suica", Display: "Suica", Flag: "🇨🇭"},
	}},
	{Name: "Grupo C", Teams: []Team{
		{Name: "Brasil", Display: "Brasil", Flag: "🇧🇷"},
		{Name: "Marrocos", Display: "Marrocos", Flag: "🇲🇦"},
		{Name: "Haiti", Display: "Haiti", Flag: "🇭🇹"},
		{Name: "Escocia", Display: "Escocia", Flag: "🏴"},
	}},
	{Name: "Grupo D", Teams: []Team{
		{Name: "Estados Unidos", Display: "Estados Unidos", Flag: "🇺🇸"},
		{Name: "Paraguai", Display: "Paraguai", Flag: "🇵🇾"},
		{Name: "Australia", Display: "Australia", Flag: "🇦🇺"},
		{Name: "Turquia", Display: "Turquia", Flag: "🇹🇷"},
	}},
	{Name: "Grupo E", Teams: []Team{
		{Name: "Alemanha", Display: "Alemanha", Flag: "🇩🇪"},
		{Name: "Curacao", Display: "Curacao", Flag: "🇨🇼"},
		{Name: "Costa do Marfim", Display: "Costa do Marfim", Flag: "🇨🇮"},
		{Name: "Equador", Display: "Equador", Flag: "🇪🇨"},
	}},
	{Name: "Grupo F", Teams: []Team{
		{Name: "Holanda", Display: "Paises Baixos", Flag: "🇳🇱"},
		{Name: "Japao", Display: "Japao", Flag: "🇯🇵"},
		{Name: "Suecia", Display: "Suecia", Flag: "🇸🇪"},
		{Name: "Tunisia", Display: "Tunisia", Flag: "🇹🇳"},
	}},
	{Name: "Grupo G", Teams: []Team{
		{Name: "Belgica", Display: "Belgica", Flag: "🇧🇪"},
		{Name: "Egito", Display: "Egito", Flag: "🇪🇬"},
		{Name: "Ira", Display: "Ira", Flag: "🇮🇷"},
		{Name: "Nova Zelandia", Display: "Nova Zelandia", Flag: "🇳🇿"},
	}},
	{Name: "Grupo H", Teams: []Team{
		{Name: "Espanha", Display: "Espanha", Flag: "🇪🇸"},
		{Name: "Cabo Verde", Display: "Cabo Verde", Flag: "🇨🇻"},
		{Name: "Arabia Saudita", Display: "Arabia Saudita", Flag: "🇸🇦"},
		{Name: "Uruguai", Display: "Uruguai", Flag: "🇺🇾"},
	}},
	{Name: "Grupo I", Teams: []Team{
		{Name: "Franca", Display: "Franca", Flag: "🇫🇷"},
		{Name: "Senegal", Display: "Senegal", Flag: "🇸🇳"},
		{Name: "Iraque", Display: "Iraque", Flag: "🇮🇶"},
		{Name: "Noruega", Display: "Noruega", Flag: "🇳🇴"},
	}},
	{Name: "Grupo J", Teams: []Team{
		{Name: "Argentina", Display: "Argentina", Flag: "🇦🇷"},
		{Name: "Argelia", Display: "Argelia", Flag: "🇩🇿"},
		{Name: "Austria", Display: "Austria", Flag: "🇦🇹"},
		{Name: "Jordania", Display: "Jordania", Flag: "🇯🇴"},
	}},
	{Name: "Grupo K", Teams: []Team{
		{Name: "Portugal", Display: "Portugal", Flag: "🇵🇹"},
		{Name: "RD Congo", Display: "RD Congo", Flag: "🇨🇩"},
		{Name: "Uzbequistao", Display: "Uzbequistao", Flag: "🇺🇿"},
		{Name: "Colombia", Display: "Colombia", Flag: "🇨🇴"},
	}},
	{Name: "Grupo L", Teams: []Team{
		{Name: "Inglaterra", Display: "Inglaterra", Flag: "🏴"},
		{Name: "Croacia", Display: "Croacia", Flag: "🇭🇷"},
		{Name: "Gana", Display: "Gana", Flag: "🇬🇭"},
		{Name: "Panama", Display: "Panama", Flag: "🇵🇦"},
	}},
}

func TeamNames() []string {
	seen := make(map[string]bool)
	var names []string
	for _, group := range Groups {
		for _, team := range group.Teams {
			if seen[team.Name] {
				continue
			}
			seen[team.Name] = true
			names = append(names, team.Name)
		}
	}
	return names
}

func ValidTeam(name string) bool {
	for _, team := range TeamNames() {
		if team == name {
			return true
		}
	}
	return false
}

func ValidGroupTeam(groupName, teamName string) bool {
	for _, group := range Groups {
		if group.Name != groupName {
			continue
		}
		for _, team := range group.Teams {
			if team.Name == teamName {
				return true
			}
		}
	}
	return false
}

func TeamByName(name string) (Team, bool) {
	for _, group := range Groups {
		for _, team := range group.Teams {
			if team.Name == name {
				return team, true
			}
		}
	}
	return Team{}, false
}

func TeamDisplay(name string) string {
	team, ok := TeamByName(name)
	if !ok {
		return name
	}
	if team.Display != "" {
		return team.Display
	}
	return team.Name
}
