package servicenow

// Schedule
/*
{'result':
    [
        {
            'name': 'GCS OnCall CCloud Compute Virtual APJ Weekend',
            'sys_id': '161cd......4e0501310f',
            'group_sys_id': '3a54562.....3961906'},
        {
            ....
*/
//ScheduleObject
type Schedule struct {
	Shifts   []ScheduleShift
	Members  []Member
	OnOnCall []Member
	ID       string
	GroupID  string
}


type LinkObj struct {
    Link string `json:"link"`
    Value string `json:"value"`
}

//TicketObject
type Ticket struct {
    SysUpdateOn  string `json:"sys_updated_on"`
    Number string `json:"number"`
    Priority int `json:"priority"`
    ShortDescription string `json:"short_description"`
    SysClassName  string `json:"sys_class_name"`
    AssignedTo  string `json:"assigned_to"`
    ClosedAt string `json:"closed_at"`
    OpenedAt  string `json:"opened_at"`
    AssignmentGroupLink LinkObj `json:"assignment_group"`
    Description string `json:"description"`//: "<br />Bridge Call URL: <a href=\"https://teams.microsoft.com/l/meetup-join/19%3ameeting_YjhhNTcwNDctNTJiMi00NWRiLWI1MjktMzNiYTRkZGVkYjAx%40thread.v2/0?context&#61;%7b%22Tid%22%3a%2242f7676c-f455-423c-82f6-dc2d99791af7%22%2c%22Oid%22%3a%221825333a-0083-4557-8827-e819758489aa%22%7d\" target=\"_blank\" rel=\"nofollow noopener noreferrer\">https://teams.microsoft.com/l/meetup-join/19%3ameeting_YjhhNTcwNDctNTJiMi00NWRiLWI1MjktMzNiYTRkZGVkYjAx%40thread.v2/0?context&#61;%7b%22Tid%22%3a%2242f7676c-f455-423c-82f6-dc2d99791af7%22%2c%22Oid%22%3a%221825333a-0083-4557-8827-e819758489aa%22%7d</a>",
    SysId string `json:"sys_id"`
    Urgency int `json:"urgency"`
    UType string `json:"u_type"`
    IncidentLink LinkObj `json:"incident"`
}

/*
{'result':

	[
	    {
	        'name': 'GCS OnCall ... APJ Weekend',
	        'sys_id': '161cd908c...4dfbf04e0501310f',
	        'group_sys_id': '3a545625d...50034ca8ebd3961906'},
	        ...
*/
type ScheduleShifts struct {
	Shifts []ScheduleShift `json:"result"`
}
type ScheduleShift struct {
	Name    string `json:"name"`
	ID      string `json:"sys_id"`
	GroupID string `json:"group_sys_id"`
}

/*
{'result':

	[{
	    'name': 'Axxxxx Axxxxx (C5xxx)',
	    'sys_id': '5d4b0466d...082154cb1159619b0',
	    'user_email': '',
	    'user_contact_number': '',
	    'userID': '5d4b0466db8...54cb1159619b0',
	    'initials': 'AA',
	    'avatar': 'images/profile/buddy_default.pngx'},
	..
*/
type Members struct {
	Members []Member `json:"result"`
}
type Member struct {
	Name              string `json:"name"`
	ID                string `json:"sys_id"`
	Mail              string `json:"user_email"`
	Contact           string `json:"user_contact_number"`
	UserID            string `json:"userID"`
	Intitials         string `json:"initials"`
	Avatar            string `json:"avatar"`
	SlackDisplayValue string
}

/*
	[{
		'memberId': 'bfc22...436d43b0',
		'memberIds': [],
		'userId': '40fb77d...03dd4bcb6d',
		'userIds': [],
		'roster': 'd61adfc2..d436d4321',
		'rota': 'db29570a47e7..436d43b1',
		'group': '76edfeb7db..a8ebd3961968',
		'escalationGroups': [],
		'deviceId': '',
		'deviceIds': [],
		'isDevice': False,
		'order': 1.0,
		'isOverride': True,
		'rotationScheduleId': '1b29570a4..3b2',
		'memberScheduleId': '3bc22346..62b7cd436d43b3'}
*/
type WhoIsOnOnCallObjects struct {
	Members []WhoIsOnOnCallObject `json:"result"`
}
type WhoIsOnOnCallObject struct {
	MemberID string `json:"memberId"`
	//MemberIds          []string `json:"memberIds"`
	UserID string `json:"userId"`
	//UserIds            []string `json:"userIds"`
	Roster string `json:"roster"`
	Rota   string `json:"rota"`
	Group  string `json:"group"`
	//EscalationGroups   []string `json:"escalationGroups"`
	DeviceID string `json:"deviceId"`
	//DeviceIds          []string `json:"deviceIds"`
	IsDevice           bool    `json:"isDevice"`
	Order              float32 `json:"order"`
	IsOverride         bool    `json:"isOverride"`
	RotationScheduleID string  `json:"rotationScheduleId"`
	MemberScheduleID   string  `json:"memberScheduleId"`
}

/*
{"current_date_time":"2023-12-23 23:39:17",
"spans":[

	{
		"color": "#FACFD7",
		"end": "2023-12-23 23:59:59",
		"group_id": "e47b17681b6aa9505f03dcef9b4bcbc4,125bb5c01bcbc150d9c921fbbb4bcb27,e479a92b1b989150341e11739b4bcb51,594e3b591baad5948ff341939b4bcb60,ce8e9823dba384143da8366af4961990,14999a71db5550103da8366af4961980,3a545625dbda2d50034ca8ebd3961906,5a3afd8c1b8bc150d9c921fbbb4bcb87,0592dcd4db20bf804e94838405961942,6b7316851b976410d9c921fbbb4bcbdf",
		"id": "100_cal_item_cmn_rota_member_3125813b47a721540262b7cd436d433e_cmn_rota_roster_96602d4cc3b2e9504dfbf04e05013171_20231223170000_20231223235959",
		"roster_id": "96602d4cc3b2e9504dfbf04e05013171",
		"rota_id": "a21cd908c37ea9504dfbf04e0501311a",
		"start": "2023-12-23 17:00:00",
		"sys_id": "3125813b47a721540262b7cd436d433e",
		"table": "cmn_rota_member",
		"textColor": "#000000",
		"title": "Himani Pathak (I568718) (Primary)",
		"user_id": "489bfee3dbba459cf4590bf5f3961920"
	},
*/
type Spans struct {
	Spans           []Span `json:"spans"`
	CurrentDateTime string `json:"current_date_time"`
}
type Span struct {
	Color     string `json:"color"`
	End       string `json:"end"`
	GroupID   string `json:"group_id"`
	ID        string `json:"id"`
	RosterID  string `json:"roster_id"`
	RotaID    string `json:"rota_id"`
	Start     string `json:"start"`
	SysID     string `json:"sys_id"`
	Table     string `json:"table"`
	TextColor string `json:"textColor"`
	Title     string `json:"title"` //* : "Firstname Familyname (Ixxxxx) (Primary)",
	UserID    string `json:"user_id"`
}
