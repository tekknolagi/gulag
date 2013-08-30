package gulag

import (
	"time"
	"sync
	"io/ioutil" 
)

type Thread struct {
	title		string
	body		string
	id		uint64
	pt		Time	/* Time of post. */
	tags		Tag[]
	posts		string[]
	next, prev	*Thread
}

var first_post	*Thread
var bump_lock	sync.Mutex
var qlookup	map[uint64]

const MAX_POSTS = 30

/*
 * Bump a thread without any posts in it. This just creates the post,
 * essentially.
 */
func (p *Thread) Bump() {
	if first_post == p {
		/* IDK what I'm doing, but this is clearly wrong. */
		return
	}
	bump_lock.Lock()
	qlookup[p.id] = p
	p.prev = nil
	p.next = first_post
	first_post = p
	bump_lock.Unlock()
}	

/*
 * Bump a thread by adding a post to it. The post should already be processed
 * HTML.
 */
func (p *Thread) Bump(string post) {
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
		for (i, cp) := (0, first_post); i < MAX_POSTS; i++ {
			if cp.next == nil {
				break
			}
			cp = cp.next
		}
		if cp.next != nil {
			/*
			 * Store the file on disk. It now no longer can be
			 * posted to (change that in the future?).
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
				err = ioutil.WriteFile(string(np.id), b)
				if err != nil {
					fmt.Println(err)
					return
				}
			}			
		}
					
	}
	bump_lock.Unlock()
}

