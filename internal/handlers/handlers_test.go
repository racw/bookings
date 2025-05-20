package handlers

import (
	"context"
	"encoding/json"
	// "fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"

	// "net/url"
	"strings"
	"testing"

	// "time"

	"github.com/racw/bookings/internal/driver"
	"github.com/racw/bookings/internal/models"
)

type postData struct {
	key   string
	value string
}

var theTests = []struct {
	name               string
	url                string
	method             string
	expectedStatusCode int
}{
	{"home", "/", "GET", http.StatusOK},
	{"about", "/about", "GET", http.StatusOK},
	{"generals-quarters", "/generals-quarters", "GET", http.StatusOK},
	{"majors-suite", "/majors-suite", "GET", http.StatusOK},
	{"search-availability", "/search-availability", "GET", http.StatusOK},
	{"contact", "/contact", "GET", http.StatusOK},
	{"non-existent-page", "/foo/bar/egg/ham", "GET", http.StatusNotFound},
	// new routes
	{"login", "/user/login", "GET", http.StatusOK},
	{"logout", "/user/logout", "GET", http.StatusOK},
	{"dashboard", "/admin/dashboard", "GET", http.StatusOK},
	{"new res", "/admin/reservations-new", "GET", http.StatusOK},
	{"all res", "/admin/reservations-all", "GET", http.StatusOK},
	{"show res", "/admin/reservations/new/1/show", "GET", http.StatusOK},
	{"show res cal", "/admin/reservations-calendar", "GET", http.StatusOK},
	{"show res cal with params", "/admin/reservations-calendar?y=2025&m=5", "GET", http.StatusOK},

	// {"post-search", "/search-availability", "POST", []postData{
	// 	{key: "start", value: "2020-01-01"},
	// 	{key: "end", value: "2020-01-02"},
	// }, http.StatusOK},
	// {"post-search-json", "/search-availability-json", "POST", []postData{
	// 	{key: "start", value: "2020-01-01"},
	// 	{key: "end", value: "2020-01-02"},
	// }, http.StatusOK},
	// {"post-reservation", "/make-reservation", "POST", []postData{
	// 	{key: "first_name", value: "John"},
	// 	{key: "last_name", value: "Smith"},
	// 	{key: "email", value: "me@here.com"},
	// 	{key: "phone", value: "123-456-7890"},
	// }, http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	for _, e := range theTests {
		resp, err := ts.Client().Get(ts.URL + e.url)
		if err != nil {
			t.Log(err)
			t.Fatal(err)
		}
		if resp.StatusCode != e.expectedStatusCode {
			t.Errorf("for %s, expected %d but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
		}
	}
}

// data for Reservation handler, /make-reservation route
var reservationTests = []struct {
	name               string
	reservation        models.Reservation
	expectedStatusCode int
	expectedLocation   string
	expectedHTML       string
}{
	{
		name: "reservation-in-session",
		reservation: models.Reservation{
			RoomID: 1,
			Room: models.Room{
				ID:       1,
				RoomName: "General's Quarters",
			},
		},
		expectedStatusCode: http.StatusOK,
		expectedHTML:       `action="/make-reservation"`,
	},
	{
		name: "reservation-not-in-session",
		reservation: models.Reservation{},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
		expectedHTML:       "",
	},
	{
		name: "non-existent-room",
		reservation: models.Reservation{
			RoomID: 100,
			Room: models.Room{
				ID:       100,
				RoomName: "General's Quarters",
			},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
		expectedHTML:       "",
	},
}

func TestRepository_Reservation(t *testing.T) {
	for _, e := range reservationTests {
		req, _ := http.NewRequest("GET", "/make-reservation", nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		if e.reservation.RoomID > 0 {
			//add reservation to session
			session.Put(ctx, "reservation", e.reservation)
		}

		handler := http.HandlerFunc(Repo.Reservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s, returned wrong response code: got %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
		}

		if e.expectedLocation != "" {
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != e.expectedLocation {
				t.Errorf("failed %s: expected location %s, but got location %s", e.name, e.expectedLocation, actualLoc.String())
			}
		}

		if e.expectedHTML != "" {
			html := rr.Body.String()
			if !strings.Contains(html, e.expectedHTML) {
				t.Errorf("failed %s: expected to find %s but did not",e.name, e.expectedHTML)
			}
		}
	}
}

type postTestCase struct {
	name               string
	url                string
	useSession         bool
	postedData         url.Values
	expectedStatusCode int
	expectedLocation   string
	expectedHTML       string
	expectedOK		   bool
	testJSON		   bool
	expectedMsg        string
}

var postReservationTests = []postTestCase{
	{
		name:              "valid-posted-data-in-session",
		url:               "/make-reservation",
		useSession:        true,
		postedData:        url.Values{
			"start_date": {"2050-01-01"},
			"end_date":   {"2050-01-02"},
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"123-456-7890"},
			"room_id":    {"1"},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/reservation-summary",
		expectedHTML:       "",
	},
	{
		name:              "missing-post-body",
		url:               "/make-reservation",
		postedData:        nil,
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
		expectedHTML:       "",
	},
	{
		name:              "invalid-start-date",
		url:               "/make-reservation",
		postedData:        url.Values{
			"start_date": {"invalid"},
			"end_date":   {"2050-01-02"},
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"123-456-7890"},
			"room_id":    {"1"},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
		expectedHTML:       "",
	},
	{
		name:              "invalid-end-date",
		url:               "/make-reservation",
		postedData:        url.Values{
			"start_date": {"2050-01-01"},
			"end_date":   {"invalid"},
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"123-456-7890"},
			"room_id":    {"1"},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
		expectedHTML:       "",
	},
	{
		name:              "invalid-room-id",
		url:               "/make-reservation",
		postedData:        url.Values{
			"start_date": {"2050-01-01"},
			"end_date":   {"2050-01-02"},
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"123-456-7890"},
			"room_id":    {"invalid"},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
		expectedHTML:       "",
	},
	{
		name:              "invalid-data",
		url:               "/make-reservation",
		useSession:        true,
		postedData:        url.Values{
			"start_date": {"2050-01-01"},
			"end_date":   {"2050-01-02"},
			"first_name": {"J"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"123-456-7890"},
			"room_id":    {"1"},
		},
		expectedStatusCode: http.StatusOK,
		expectedLocation:   "",
		expectedHTML:       `action="/make-reservation"`,
	},
	{
		name:              "database-insert-fails-reservation",
		url:               "/make-reservation",
		useSession:        true,
		postedData:        url.Values{
			"start_date": {"2050-01-01"},
			"end_date":   {"2050-01-02"},
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"123-456-7890"},
			"room_id":    {"12"},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
		expectedHTML:       "",
	},
	{
		name:              "database-insert-fails-restriction",
		url:               "/make-reservation",
		useSession:        true,
		postedData:        url.Values{
			"start_date": {"2050-01-01"},
			"end_date":   {"2050-01-02"},
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"123-456-7890"},
			"room_id":    {"1000"},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
		expectedHTML:       "",
	},
	
}

func testPostHandler(t *testing.T, testCases []postTestCase, handlerFunc http.HandlerFunc) {
	for _, e := range testCases {
		var req *http.Request
		if e.postedData != nil {
			req, _ = http.NewRequest("POST", e.url, strings.NewReader(e.postedData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", e.url, nil)
		}
	
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		if e.useSession {
			session.Put(ctx, "reservation", models.Reservation{})
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		
		handler := http.HandlerFunc(handlerFunc)
		handler.ServeHTTP(rr, req)

		if e.testJSON {
			var j jsonResponse
			err := json.Unmarshal([]byte(rr.Body.String()), &j)
			if err != nil {
				t.Error("failed to parse json!")
			}

			if j.OK != e.expectedOK {
				t.Errorf("%s: expected %v but got %v", e.name, e.expectedOK, j.OK)
			}

			if j.Message != e.expectedMsg {
				t.Errorf("Expected message: %s but got: %s",e.expectedMsg, j.Message)
			}
		} else {
			if rr.Code != e.expectedStatusCode {
				t.Errorf("%s, returned wrong response code: got %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
			}
	
			if e.expectedLocation != "" {
				actualLoc, _ := rr.Result().Location()
				if actualLoc.String()!= e.expectedLocation {
					t.Errorf("failed %s: expected location %s, but got location %s", e.name, e.expectedLocation, actualLoc.String())
				}
			}
	
			if e.expectedHTML!= "" {
				html := rr.Body.String()
				if!strings.Contains(html, e.expectedHTML) {
					t.Errorf("failed %s: expected to find %s but did not",e.name, e.expectedHTML)
				}
			}
		}
	
		
	}
}

func TestRepository_PostReservation(t *testing.T) {
	testPostHandler(t, postReservationTests, Repo.PostReservation)

}

func TestNewRepo(t *testing.T) {
	var db driver.DB
	testRepo := NewRepo(&app, &db)

	if reflect.TypeOf(testRepo).String() != "*handlers.Repository" {
		t.Errorf("NewRepo returned wrong type: got %s, want %s", reflect.TypeOf(testRepo).String(), "*handlers.Repository")
	}
}

func TestRepository_Availability(t *testing.T) {
	var testCases = []postTestCase {
		{
			name: "rooms not available",
			url: "/search-availability",
			postedData: url.Values{
				"start": {"2050-01-01"},
				"end":	 {"2050-01-02"},
			},
			expectedStatusCode: http.StatusSeeOther,
		},
		{
			name: "rooms are available",
			url: "/search-availability",
			postedData: url.Values{
				"start": {"2040-01-01"},
				"end":	 {"2040-01-02"},
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "empty post body",
			url: "/search-availability",
			postedData: url.Values{},
			expectedStatusCode: http.StatusSeeOther,
		},
		{
			name: "start date invalid",
			url: "/search-availability",
			postedData: url.Values{
				"start": {"invalid"},
				"end":	 {"2040-01-02"},
			},
			expectedStatusCode: http.StatusSeeOther,
		},
		{
			name: "end date invalid",
			url: "/search-availability",
			postedData: url.Values{
				"start": {"2040-01-01"},
				"end":	 {"invalid"},
			},
			expectedStatusCode: http.StatusSeeOther,
		},
		{
			name: "database query fails",
			url: "/search-availability",
			postedData: url.Values{
				"start": {"2060-01-01"},
				"end":	 {"2060-01-02"},
			},
			expectedStatusCode: http.StatusSeeOther,
		},
	}
	testPostHandler(t, testCases, Repo.PostAvailability)
}
func TestRepository_AvailabilityJSON(t *testing.T) {
	var jsonTestCases = []postTestCase{
		{
			name: "rooms not available",
			postedData: url.Values{
				"start": 	{ "2050-01-01"},
				"end": 		{ "2050-01-02"},
				"room_id": 	{"1"},
			},
			testJSON: true,
			expectedOK: false,
		},
		{
			name: "rooms are available",
			postedData: url.Values{
				"start": 	{ "2040-01-01"},
				"end": 		{ "2040-01-02"},
				"room_id": 	{"1"},
			},
			testJSON: true,
			expectedOK: true,
		},
		{
			name: "empty post body",
			postedData:  nil,
			testJSON: true,
			expectedOK: false,
			expectedMsg: "Internal server error",
		},
		{
			name: "database query fails",
			postedData: url.Values{
				"start": 	{ "2060-01-01"},
				"end": 		{ "2060-01-02"},
				"room_id": 	{"1"},
			},
			testJSON: true,
			expectedOK: false,
			expectedMsg: "Error querying database",
		},
	}
	testPostHandler(t, jsonTestCases, Repo.AvailabilityJSON)
}

func TestRepository_ReservationSummary(t *testing.T) {
	// 1st case reservation in session
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}
	req, _ := http.NewRequest("GET", "/reservation-summary", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)
	handler := http.HandlerFunc(Repo.ReservationSummary)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("ReservationSummary handler returned wrong status code: got %d, want %d", rr.Code, http.StatusOK)
	}
	// 2nd case reservation not in session
	req, _ = http.NewRequest("GET", "/reservation-summary", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.ReservationSummary)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("ReservationSummary handler returned wrong status code: got %d, want %d", rr.Code, http.StatusSeeOther)
	}
}

func TestRepository_ChooseRoom(t *testing.T) {
	// 1st case reservation in session
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}
	req, _ := http.NewRequest("GET", "/choose-room/1", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)
	req.RequestURI = "/choose-room/1"

	rr := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)
	handler := http.HandlerFunc(Repo.ChooseRoom)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("ChooseRoom handler returned wrong status code: got %d, want %d", rr.Code, http.StatusSeeOther)
	}

	// 2nd case reservation not in session
	req, _ = http.NewRequest("GET", "/choose-room/1", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.RequestURI = "/choose-room/1"
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.ChooseRoom)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("ChooseRoom handler returned wrong status code: got %d, want %d", rr.Code, http.StatusSeeOther)
	}

	// 3rd case invalid room id or missing url parameter
	req, _ = http.NewRequest("GET", "/choose-room/fish", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.RequestURI = "/choose-room/fish"

	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.ChooseRoom)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("ChooseRoom handler returned wrong status code: got %d, want %d", rr.Code, http.StatusSeeOther)
	}
}

func TestRepository_BookRoom(t *testing.T) {
	// 1st case database works
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}
	req, _ := http.NewRequest("GET", "/book-room?s=2050-01-01&e=2050-01-02&id=1", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)
	handler := http.HandlerFunc(Repo.BookRoom)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("BookRoom handler returned wrong status code: got %d, want %d", rr.Code, http.StatusSeeOther)
	}
	// 2nd case database fails
	req, _ = http.NewRequest("GET", "/book-room?s=2040-01-01&e=2040-01-02&id=4", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.BookRoom)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("BookRoom handler returned wrong status code: got %d, want %d", rr.Code, http.StatusSeeOther)
	}
}

var loginTests = []struct {
	name               string
	email              string
	expectedStatusCode int
	expectedHTML       string
	expectedLocation   string
}{
	{
		"valid-credentials",
		"me@here.ca",
		http.StatusSeeOther,
		"",
		"/",
	},
	{
		"invalid-credentials",
		"jack@invalid.com",
		http.StatusSeeOther,
		"",
		"/user/login",
	},
	{
		"invalid-data",
		"j",
		http.StatusOK,
		`action="/user/login"`,
		"",
	},
}

func TestLogin(t *testing.T) {
	// range through all tests
	for _, e := range loginTests {
		postedData := url.Values{}
		postedData.Add("email", e.email)
		postedData.Add("password", "password")

		//create request
		req, _ := http.NewRequest("POST", "/user/login", strings.NewReader(postedData.Encode()))
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		// set the header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		// create handler
		handler := http.HandlerFunc(Repo.PostShowLogin)
		handler.ServeHTTP(rr, req)

		// check the status code
		if rr.Code != e.expectedStatusCode {
			t.Errorf("failed %s: expected code %d, but got %d", e.name, e.expectedStatusCode, rr.Code)
		}

		if e.expectedLocation != "" {
			// get URL from test
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != e.expectedLocation {
				t.Errorf("failed %s: expected location %s, but got %s", e.name, e.expectedLocation, actualLoc.String())
			}
		}

		//checking for expected values in HTML
		if e.expectedHTML != "" {
			// check the body
			html := rr.Body.String()
			if !strings.Contains(html, e.expectedHTML) {
				t.Errorf("failed %s: expected to find %s, but did not", e.name, e.expectedHTML)
			}
		}
	}
}

func getCtx(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}
	return ctx
}
