package copa

type FixtureSeed struct {
	Round     int
	Date      string
	GroupName string
	HomeTeam  string
	AwayTeam  string
}

type KnockoutFixtureSeed struct {
	Round    int
	DateTime string
	HomeTeam string
	AwayTeam string
}

var KnockoutFixtures = []KnockoutFixtureSeed{
	{Round: 4, DateTime: "2026-06-28T16:00", HomeTeam: "Africa do Sul", AwayTeam: "Canada"},
	{Round: 4, DateTime: "2026-06-29T14:00", HomeTeam: "Brasil", AwayTeam: "Japao"},
	{Round: 4, DateTime: "2026-06-29T17:30", HomeTeam: "Alemanha", AwayTeam: "Paraguai"},
	{Round: 4, DateTime: "2026-06-29T22:00", HomeTeam: "Holanda", AwayTeam: "Marrocos"},
	{Round: 4, DateTime: "2026-06-30T14:00", HomeTeam: "Costa do Marfim", AwayTeam: "Noruega"},
	{Round: 4, DateTime: "2026-06-30T18:00", HomeTeam: "Franca", AwayTeam: "Suecia"},
	{Round: 4, DateTime: "2026-06-30T22:00", HomeTeam: "Mexico", AwayTeam: "Equador"},
	{Round: 4, DateTime: "2026-07-01T13:00", HomeTeam: "Inglaterra", AwayTeam: "RD Congo"},
	{Round: 4, DateTime: "2026-07-01T17:00", HomeTeam: "Belgica", AwayTeam: "Senegal"},
	{Round: 4, DateTime: "2026-07-01T21:00", HomeTeam: "Estados Unidos", AwayTeam: "Bosnia"},
	{Round: 4, DateTime: "2026-07-02T16:00", HomeTeam: "Espanha", AwayTeam: "Austria"},
	{Round: 4, DateTime: "2026-07-02T20:00", HomeTeam: "Portugal", AwayTeam: "Croacia"},
	{Round: 4, DateTime: "2026-07-03T00:00", HomeTeam: "Suica", AwayTeam: "Argelia"},
	{Round: 4, DateTime: "2026-07-03T15:00", HomeTeam: "Australia", AwayTeam: "Egito"},
	{Round: 4, DateTime: "2026-07-03T19:00", HomeTeam: "Argentina", AwayTeam: "Cabo Verde"},
	{Round: 4, DateTime: "2026-07-03T22:30", HomeTeam: "Colombia", AwayTeam: "Gana"},

	{Round: 5, DateTime: "2026-07-04T14:00", HomeTeam: "Canada", AwayTeam: "Marrocos"},
	{Round: 5, DateTime: "2026-07-04T18:00", HomeTeam: "Paraguai", AwayTeam: "Franca"},
	{Round: 5, DateTime: "2026-07-05T17:00", HomeTeam: "Brasil", AwayTeam: "Noruega"},
	{Round: 5, DateTime: "2026-07-05T21:00", HomeTeam: "Mexico", AwayTeam: "Inglaterra"},
	{Round: 5, DateTime: "2026-07-06T16:00", HomeTeam: "Portugal", AwayTeam: "Espanha"},
	{Round: 5, DateTime: "2026-07-06T21:00", HomeTeam: "Estados Unidos", AwayTeam: "Belgica"},
	{Round: 5, DateTime: "2026-07-07T13:00", HomeTeam: "Argentina", AwayTeam: "Egito"},
	{Round: 5, DateTime: "2026-07-07T17:00", HomeTeam: "Suica", AwayTeam: "Colombia"},
}

var GroupStageFixtures = []FixtureSeed{
	{Round: 1, Date: "2026-06-11", GroupName: "Grupo A", HomeTeam: "Mexico", AwayTeam: "Africa do Sul"},
	{Round: 1, Date: "2026-06-11", GroupName: "Grupo A", HomeTeam: "Coreia do Sul", AwayTeam: "Republica Tcheca"},
	{Round: 1, Date: "2026-06-12", GroupName: "Grupo B", HomeTeam: "Canada", AwayTeam: "Bosnia"},
	{Round: 1, Date: "2026-06-12", GroupName: "Grupo D", HomeTeam: "Estados Unidos", AwayTeam: "Paraguai"},
	{Round: 1, Date: "2026-06-13", GroupName: "Grupo B", HomeTeam: "Catar", AwayTeam: "Suica"},
	{Round: 1, Date: "2026-06-13", GroupName: "Grupo C", HomeTeam: "Brasil", AwayTeam: "Marrocos"},
	{Round: 1, Date: "2026-06-13", GroupName: "Grupo C", HomeTeam: "Haiti", AwayTeam: "Escocia"},
	{Round: 1, Date: "2026-06-13", GroupName: "Grupo D", HomeTeam: "Australia", AwayTeam: "Turquia"},
	{Round: 1, Date: "2026-06-14", GroupName: "Grupo E", HomeTeam: "Alemanha", AwayTeam: "Curacao"},
	{Round: 1, Date: "2026-06-14", GroupName: "Grupo E", HomeTeam: "Costa do Marfim", AwayTeam: "Equador"},
	{Round: 1, Date: "2026-06-14", GroupName: "Grupo F", HomeTeam: "Holanda", AwayTeam: "Japao"},
	{Round: 1, Date: "2026-06-14", GroupName: "Grupo F", HomeTeam: "Suecia", AwayTeam: "Tunisia"},
	{Round: 1, Date: "2026-06-15", GroupName: "Grupo G", HomeTeam: "Belgica", AwayTeam: "Egito"},
	{Round: 1, Date: "2026-06-15", GroupName: "Grupo G", HomeTeam: "Ira", AwayTeam: "Nova Zelandia"},
	{Round: 1, Date: "2026-06-15", GroupName: "Grupo H", HomeTeam: "Espanha", AwayTeam: "Cabo Verde"},
	{Round: 1, Date: "2026-06-15", GroupName: "Grupo H", HomeTeam: "Arabia Saudita", AwayTeam: "Uruguai"},
	{Round: 1, Date: "2026-06-16", GroupName: "Grupo I", HomeTeam: "Franca", AwayTeam: "Senegal"},
	{Round: 1, Date: "2026-06-16", GroupName: "Grupo I", HomeTeam: "Iraque", AwayTeam: "Noruega"},
	{Round: 1, Date: "2026-06-16", GroupName: "Grupo J", HomeTeam: "Argentina", AwayTeam: "Argelia"},
	{Round: 1, Date: "2026-06-16", GroupName: "Grupo J", HomeTeam: "Austria", AwayTeam: "Jordania"},
	{Round: 1, Date: "2026-06-17", GroupName: "Grupo K", HomeTeam: "Portugal", AwayTeam: "RD Congo"},
	{Round: 1, Date: "2026-06-17", GroupName: "Grupo K", HomeTeam: "Uzbequistao", AwayTeam: "Colombia"},
	{Round: 1, Date: "2026-06-17", GroupName: "Grupo L", HomeTeam: "Inglaterra", AwayTeam: "Croacia"},
	{Round: 1, Date: "2026-06-17", GroupName: "Grupo L", HomeTeam: "Gana", AwayTeam: "Panama"},
	{Round: 2, Date: "2026-06-18", GroupName: "Grupo A", HomeTeam: "Republica Tcheca", AwayTeam: "Africa do Sul"},
	{Round: 2, Date: "2026-06-18", GroupName: "Grupo A", HomeTeam: "Mexico", AwayTeam: "Coreia do Sul"},
	{Round: 2, Date: "2026-06-18", GroupName: "Grupo B", HomeTeam: "Suica", AwayTeam: "Bosnia"},
	{Round: 2, Date: "2026-06-18", GroupName: "Grupo B", HomeTeam: "Canada", AwayTeam: "Catar"},
	{Round: 2, Date: "2026-06-19", GroupName: "Grupo C", HomeTeam: "Escocia", AwayTeam: "Marrocos"},
	{Round: 2, Date: "2026-06-19", GroupName: "Grupo C", HomeTeam: "Brasil", AwayTeam: "Haiti"},
	{Round: 2, Date: "2026-06-19", GroupName: "Grupo D", HomeTeam: "Estados Unidos", AwayTeam: "Australia"},
	{Round: 2, Date: "2026-06-19", GroupName: "Grupo D", HomeTeam: "Turquia", AwayTeam: "Paraguai"},
	{Round: 2, Date: "2026-06-20", GroupName: "Grupo E", HomeTeam: "Alemanha", AwayTeam: "Costa do Marfim"},
	{Round: 2, Date: "2026-06-20", GroupName: "Grupo E", HomeTeam: "Equador", AwayTeam: "Curacao"},
	{Round: 2, Date: "2026-06-20", GroupName: "Grupo F", HomeTeam: "Holanda", AwayTeam: "Suecia"},
	{Round: 2, Date: "2026-06-20", GroupName: "Grupo F", HomeTeam: "Tunisia", AwayTeam: "Japao"},
	{Round: 2, Date: "2026-06-21", GroupName: "Grupo G", HomeTeam: "Belgica", AwayTeam: "Ira"},
	{Round: 2, Date: "2026-06-21", GroupName: "Grupo G", HomeTeam: "Nova Zelandia", AwayTeam: "Egito"},
	{Round: 2, Date: "2026-06-21", GroupName: "Grupo H", HomeTeam: "Espanha", AwayTeam: "Arabia Saudita"},
	{Round: 2, Date: "2026-06-21", GroupName: "Grupo H", HomeTeam: "Uruguai", AwayTeam: "Cabo Verde"},
	{Round: 2, Date: "2026-06-22", GroupName: "Grupo I", HomeTeam: "Franca", AwayTeam: "Iraque"},
	{Round: 2, Date: "2026-06-22", GroupName: "Grupo I", HomeTeam: "Noruega", AwayTeam: "Senegal"},
	{Round: 2, Date: "2026-06-22", GroupName: "Grupo J", HomeTeam: "Argentina", AwayTeam: "Austria"},
	{Round: 2, Date: "2026-06-22", GroupName: "Grupo J", HomeTeam: "Jordania", AwayTeam: "Argelia"},
	{Round: 2, Date: "2026-06-23", GroupName: "Grupo K", HomeTeam: "Portugal", AwayTeam: "Uzbequistao"},
	{Round: 2, Date: "2026-06-23", GroupName: "Grupo K", HomeTeam: "Colombia", AwayTeam: "RD Congo"},
	{Round: 2, Date: "2026-06-23", GroupName: "Grupo L", HomeTeam: "Inglaterra", AwayTeam: "Gana"},
	{Round: 2, Date: "2026-06-23", GroupName: "Grupo L", HomeTeam: "Panama", AwayTeam: "Croacia"},
	{Round: 3, Date: "2026-06-24", GroupName: "Grupo B", HomeTeam: "Suica", AwayTeam: "Canada"},
	{Round: 3, Date: "2026-06-24", GroupName: "Grupo B", HomeTeam: "Bosnia", AwayTeam: "Catar"},
	{Round: 3, Date: "2026-06-24", GroupName: "Grupo A", HomeTeam: "Republica Tcheca", AwayTeam: "Mexico"},
	{Round: 3, Date: "2026-06-24", GroupName: "Grupo A", HomeTeam: "Africa do Sul", AwayTeam: "Coreia do Sul"},
	{Round: 3, Date: "2026-06-24", GroupName: "Grupo C", HomeTeam: "Escocia", AwayTeam: "Brasil"},
	{Round: 3, Date: "2026-06-24", GroupName: "Grupo C", HomeTeam: "Marrocos", AwayTeam: "Haiti"},
	{Round: 3, Date: "2026-06-25", GroupName: "Grupo E", HomeTeam: "Curacao", AwayTeam: "Costa do Marfim"},
	{Round: 3, Date: "2026-06-25", GroupName: "Grupo E", HomeTeam: "Equador", AwayTeam: "Alemanha"},
	{Round: 3, Date: "2026-06-25", GroupName: "Grupo F", HomeTeam: "Japao", AwayTeam: "Suecia"},
	{Round: 3, Date: "2026-06-25", GroupName: "Grupo F", HomeTeam: "Tunisia", AwayTeam: "Holanda"},
	{Round: 3, Date: "2026-06-25", GroupName: "Grupo D", HomeTeam: "Turquia", AwayTeam: "Estados Unidos"},
	{Round: 3, Date: "2026-06-25", GroupName: "Grupo D", HomeTeam: "Paraguai", AwayTeam: "Australia"},
	{Round: 3, Date: "2026-06-26", GroupName: "Grupo G", HomeTeam: "Egito", AwayTeam: "Ira"},
	{Round: 3, Date: "2026-06-26", GroupName: "Grupo G", HomeTeam: "Nova Zelandia", AwayTeam: "Belgica"},
	{Round: 3, Date: "2026-06-26", GroupName: "Grupo H", HomeTeam: "Cabo Verde", AwayTeam: "Arabia Saudita"},
	{Round: 3, Date: "2026-06-26", GroupName: "Grupo H", HomeTeam: "Uruguai", AwayTeam: "Espanha"},
	{Round: 3, Date: "2026-06-26", GroupName: "Grupo I", HomeTeam: "Noruega", AwayTeam: "Franca"},
	{Round: 3, Date: "2026-06-26", GroupName: "Grupo I", HomeTeam: "Senegal", AwayTeam: "Iraque"},
	{Round: 3, Date: "2026-06-27", GroupName: "Grupo J", HomeTeam: "Argelia", AwayTeam: "Austria"},
	{Round: 3, Date: "2026-06-27", GroupName: "Grupo J", HomeTeam: "Jordania", AwayTeam: "Argentina"},
	{Round: 3, Date: "2026-06-27", GroupName: "Grupo K", HomeTeam: "Colombia", AwayTeam: "Portugal"},
	{Round: 3, Date: "2026-06-27", GroupName: "Grupo K", HomeTeam: "RD Congo", AwayTeam: "Uzbequistao"},
	{Round: 3, Date: "2026-06-27", GroupName: "Grupo L", HomeTeam: "Panama", AwayTeam: "Inglaterra"},
	{Round: 3, Date: "2026-06-27", GroupName: "Grupo L", HomeTeam: "Croacia", AwayTeam: "Gana"},
}
