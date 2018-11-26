package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"regexp"

	_ "github.com/go-sql-driver/mysql"
)

// Base - This base struct
type Base struct {
	/*
	Bese struct -this is struct comment
	*/
	db *sql.DB
}

type celRecord struct {
	id, eventtype, eventtime, cidName, cidNum, cidAni, cidRdnis,
	cidDnid, exten, context, channame, appname, appdata, amaflags,
	accountcode, uniqueid, linkedid, peer, userdeftype, extra string
}

type query struct {
	Before   string `json:"Before"`
	After    string `json:"After"`
	PhoneNum string `json:"phoneNum"`
}

func (base *Base) connect(connString string) {
	db, err := sql.Open("mysql", connString)
	if err != nil {
		fmt.Println(err)
		return
	}
	base.db = db
}

func (base *Base) queryAutocomplte(filter string) *sql.Rows {
	row, err := base.db.Query(`select distinct src, max(calldate)
	from cdr
    where disposition <> 'ANSWER' and src like '%` + filter + `%'
    group by src
    order by max(calldate) desc, src
    limit 5;`)
	fmt.Println("query: ", `select distinct src, max(calldate)
	from cdr
    where disposition <> 'ANSWER' and src like '%` + filter + `%'
    group by src
    order by max(calldate) desc, src
    limit 5;`)
	if err != nil {
		fmt.Println("err: ", err)
		return nil
	}
	return row
}

var aster = Base{}
var base = Base{}

func mainHandleRoute(w http.ResponseWriter, r *http.Request) {
	strs := strings.Split(r.URL.Path, "/")
	//fmt.Println("r: ", r.URL.Parse(ref string))
	rows := base.queryAutocomplte(strs[len(strs)-1])
	fmt.Println(strs[len(strs)-1])
	complete := "["
	for rows.Next() {
		var src string
		var date string
		rows.Scan(&src, &date)
		complete += "\"" + src + "\"" + ","
	}
	complete = complete[:len(complete)-1] + "]"
	w.WriteHeader(200)
	w.Write([]byte(complete))
}

func ivrToName(s string) string {
	//fmt.Println("ivrToName")
	q := "select name from ivr_details where id ='" + s +"';"
	fmt.Println("ivrToName: ", q)
	row, err := aster.db.Query(q)
	if err != nil {
		fmt.Println("ivrToNameErr: ", err)
		return ""
	}
	ivr:=""
	for row.Next() {
		row.Scan(&ivr)
	}
	fmt.Println("ivrToName: ", ivr)
	return ivr
}

func number(w http.ResponseWriter, r *http.Request) {

	reg, _ := regexp.Compile(`ivr-([\d]*)`)
	//w.Write([]byte("ok"))
	fmt.Println("Number Header: ", r.Header)
	fmt.Println()

	b, err := ioutil.ReadAll(r.Body)
	var Query query
	err = json.Unmarshal(b, &Query)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	q1 := "select uniqueid, src, min(calldate) from cdr where calldate between '" + Query.Before + "' and '" +
		Query.After + "' and src like '%" + Query.PhoneNum + "' group by uniqueid, src"
	q2 := strings.Replace(q1, "src", "dst", -1)
	q3 := strings.Replace(q1, "src", "did", -1)
	qall := q1 + " union " + q2 + " union " + q3 + ";"
	fmt.Printf("Number qall: %s\n", qall)
	row, err := base.db.Query(qall)
	if err != nil {
		fmt.Println("err: ", err)
		w.Write([]byte(err.Error()))
		w.WriteHeader(500)
		return
	}
	html := "<div id=\"accordion\">"
	for row.Next() {
		var uniqueid, src, calldate string
		//fmt.Println("row: ", row)
		row.Scan(&uniqueid, &src, &calldate)
		cnt := "select count(*) from cel where linkedid='" + uniqueid + "';"
		c := 0
		err := base.db.QueryRow(cnt).Scan(&c)
		if err != nil {
			fmt.Println("err: ", err)
			w.Write([]byte(err.Error()))
			w.WriteHeader(500)
			return
		}
		if c>0 {
		cel := "select * from cel where linkedid='" + uniqueid + "';"
		celrow, err := base.db.Query(cel)
		if err != nil {
			fmt.Println("err: ", err)
			w.Write([]byte(err.Error()))
			w.WriteHeader(500)
			return
		}
		html += "<h3>" + src + "   " + calldate + "     " +uniqueid+"</h3>"
		html += "<div>"
		html += "<table class=\"table\"><tr><th>дата</th><th>событие</th></tr><thead></thead><tbody>"
		for celrow.Next() {
			rec := celRecord{}
			celrow.Scan(&rec.id, &rec.eventtype, &rec.eventtime,
				&rec.cidName, &rec.cidNum, &rec.cidAni, &rec.cidRdnis,
				&rec.cidDnid, &rec.exten, &rec.context, &rec.channame, &rec.appname,
				&rec.appdata, &rec.amaflags, &rec.accountcode, &rec.uniqueid,
				&rec.linkedid, &rec.peer, &rec.userdeftype, &rec.extra)
			switch rec.eventtype {
			case "CHAN_START":
				rec.eventtype = "Соединились"
			case "HANGUP":
				rec.eventtype = "Положили трубку"
			case "CHAN_END":
				rec.eventtype = "Связь закончилась"
			case "APP_START":
				switch rec.exten {
				case "recordcheck":
					continue
				}
			case "APP_END":
				switch rec.exten {
				case "recordcheck":
					continue
				}
			case "ANSWER":
				switch rec.exten {
				case "s":
					if reg.MatchString(rec.context)  {
						fmt.Println("reg.FindStringSubmatch(rec.context):", reg.FindStringSubmatch(rec.context)[1])
						rec.eventtype = "Запуск приветствия: " + ivrToName(reg.FindStringSubmatch(rec.context)[1])
					}
				}
			}
			html += "<tr><td>" + rec.eventtime + "</td><td>" + rec.eventtype + "</td></tr>"
			//html += "<p>" + rec.eventtype + "<br>" + rec.eventtime + "</p>"
		}
		html += "</table></tbody>"
		html += "</div>"
	}
	}
	html += "</div>"
	w.Write([]byte(html))
	w.WriteHeader(200)
	//fmt.Println("html: ", html)
}

func main() {
	base.connect("root:root@/asteriskcdrdb")
	aster.connect("root:root@/asterisk")
	http.HandleFunc("/number/", number)
	http.HandleFunc("/autocomplete_number/", mainHandleRoute)
	//http.HandleFunc("/autocomplete_number", autocomplete_number)
	//http.HandleFunc("/number", number)
	err := http.ListenAndServe(":6666", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
