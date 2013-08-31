package main

import (
	"fmt"
	"sync"
	"encoding/json"
	"io/ioutil"
	"math/rand"
)

type Thread struct {
	title		string
	id		uint64
//	pt		Time	/* Time of post. */
//	tags		[]Tag
	posts		[]string	/* First post is body of thread. */
	next, prev	*Thread
}

var first_post	*Thread
var bump_lock	sync.Mutex
var qlookup	map[uint64]*Thread

const MAX_POSTS = 5 // 30
const MAX_POST_LEN = 500

func init() {
	qlookup = make(map[uint64]*Thread)
}

func rand64() uint64 {
	return uint64(rand.Uint32()) << 32 | uint64(rand.Uint32())
}

/*
 * Get all the relevent posts starting from firt_post and ending at MAX_POSTS.
 * HERE IS WHERE ALL THE MAJOR BOTTLENECKS WILL BE (in terms of speed)!
 */
func getReleventThreads() []*Thread {
	posts := make([]*Thread, MAX_POSTS)
	bump_lock.Lock()
	for i, p := 0, first_post; i < MAX_POSTS; i++ {	
		if p == nil {
			break
		}
		posts[i] = p
		p = p.next
	}
	bump_lock.Unlock()
	return posts
}
	
/*
 * Bump a thread without any posts in it. This just creates the post,
 * essentially.
 */
func (p *Thread) post() {
	if first_post == p {
		/* IDK what I'm doing, but this is clearly wrong. */
		return
	}
	bump_lock.Lock()
	p.id = rand64()
	qlookup[p.id] = p
	p.prev = nil
	p.next = first_post
	if first_post != nil {
		p.next.prev = p
	}
	first_post = p
	bump_lock.Unlock()
}	

/*
 * Bump a thread by adding a post to it. The post should already be processed
 * HTML.
 */
func (p *Thread) bump(post string) {
	p.posts = append(p.posts, post)
	if first_post == p {
		/*
		 * Just return, not locking. This theoretically puts the top 
		 * post at a disadvantage of being retaining the top spot during
		 * simultaneous bumps. In reality, I highly doubt it will make a
		 * difference.
		 */
		return
	}
	bump_lock.Lock()
	if first_post == nil { /* This is the first post. */
		first_post = p
	} else {
		pp := p.prev
		pn := p.next
		p.prev = nil
		p.next = first_post
		first_post = p
		pp.next = pn
		cp := first_post
		for i := 0; i < MAX_POSTS; i++ {
			if cp.next == nil {
				break
			}
			cp = cp.next
		}
		if cp.next != nil {
			/*
			 * Store the file on disk. It now will no longer show
			 * up on the front page
			 */
			delete(qlookup, cp.next.id)
			np :=  cp.next
			cp.next = nil
			go func () {
				/* No reason to do this not concurrently. */
				b, err := json.Marshal(np)
				if err != nil {
					fmt.Println(err)
					return
				}
				err = ioutil.WriteFile(string(np.id), b, 0600)
				if err != nil {
					fmt.Println(err)
					return
				}
			}()
		}
					
	}
	bump_lock.Unlock()
}

/*
 * Lookup a thread.
 */
func getThread(id uint64) *Thread {
	var t *Thread
	var ok bool
	if t, ok = qlookup[id]; ok == false {
		/* The thread is not cached. */
		/* Read the file from the disk and return it. */
		b, err := ioutil.ReadFile(string(id))
		if err != nil {
			/* No such thread exists. */
			return nil
		}
		json.Unmarshal(b, &t)
	}
	return t
}