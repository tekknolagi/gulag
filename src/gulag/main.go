package main

import (
	"github.com/hoisie/web"
	"strconv"
	"strings"
)

const HEADER = `<!DOCTYPE html>
	<html>
	<head><title>Gulag.go - Speak freely</title></head>
	<body>`

const FOOTER = `<a href="/">Return</a></body></html>`

const ERR404 = (HEADER +
	"<h1>404!<h1><br /><b>This shit doesn't exist</b>" +
	FOOTER)

/*
 * I should really replace whatever the hell I'm doing with all the string
 * appending with something much more sane, but alas I am too tired to look up a
 * better method.
 */

func dispAllThreads() string {
	s := []string{HEADER + "<table border=\"1\">"}
	ps := getReleventThreads()
	for i := 0; i < MAX_POSTS; i++ {
		if ps[i] == nil {
			break
		}
		s = append(s, "<tr><td><a href=\"/t/" +
			strconv.FormatUint(ps[i].id, 32) +
			"\">" + ps[i].title + "</a></td></tr>")
	}
	s = append(s, `<tr><td><form action="/new" method="post">
		Title: <input type="text" name="title" value=""><br />
		Body: <input type="text" name="body" value=""><br />
		<input type="submit" value="Post"></table>` + FOOTER)
	return strings.Join(s, "")
}

func dispThreadHandler(v string) string {
	tid, err := strconv.ParseUint(v, 32, 64)
	if err != nil {
		return ERR404
	}
	t := getThread(tid)
	if t == nil {
		return ERR404
	}
	return dispThread(t, v)
}

func dispThread(t *Thread, v string) string {
	s := []string{HEADER + "<table border=\"1\"><tr><td><h1>" + t.title +
			"</h1></td></tr>"}
	for _, p := range t.posts {
		s = append(s, "<tr><td>" + p + "</tr></td>")
	}
	/* Add post form and footer. */
	s = append(s, `<tr><td><form action="/p/` + v + `" method="post">
		<input type="text" name="postValue" value="">
		<input type="submit" value="Post">
		</form></tr></td></table>` + FOOTER)
	return strings.Join(s, "")
}

func postNewThread(ctx *web.Context) string {
	title, ok := ctx.Params["title"]
	if ok == false {
		return ERR404
	}
	body, ok := ctx.Params["body"]
	if ok == false {
		return ERR404
	}
	t := &Thread{title, 0, []string{body}, nil, nil }
	t.post()
	return dispThread(t, strconv.FormatUint(t.id, 32))
}

func postToThread(ctx *web.Context, v string) string {
	tid, err := strconv.ParseUint(v, 32, 64)
	if err != nil {
		return ERR404
	}
	t := getThread(tid)
	if t == nil {
		return ERR404
	}
	pv, ok := ctx.Params["postValue"]
	if ok == false {
		return ERR404	/* This is clearly not what should happen. */
	}
	if len(pv) > MAX_POST_LEN {
		return ERR404	/* Ditto */
	}
	t.bump(pv)
	t = getThread(tid)	/* Reload the thread. */
	return dispThread(t, v)
}

func main() {
	web.Get("/", dispAllThreads)
	web.Get("/t/(.*)", dispThreadHandler)
	web.Post("/p/(.*)", postToThread)
	web.Post("/new", postNewThread)
	web.Run("0.0.0.0:9999")
}