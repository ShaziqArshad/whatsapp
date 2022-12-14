package wa

import (
	whatsapp "github.com/Rhymen/go-whatsapp"
	mgo "github.com/globalsign/mgo"
)

type ConnWrapper struct {
	IsFilled bool
	Conn     *whatsapp.Conn
	Sess     *whatsapp.Session
}

type BucketSession struct {
	Items      map[string]ConnWrapper
	MgoSession *mgo.Session
}

func (c *BucketSession) Sync() {
	// sync with existing session
	var storedSessions WASessions
	(&SessionStorage{MgoSession: c.MgoSession}).FetchAll(&storedSessions)

	if len(storedSessions) > 0 {
		for _, item := range storedSessions {
			c.Items[item.number] = ConnWrapper{
				IsFilled: true,
				Conn:     nil,
				Sess:     &item.session,
			}
		}
	}
}

func (c *BucketSession) Save(number string, conn *whatsapp.Conn, sess *whatsapp.Session) {
	c.Items[number] = ConnWrapper{
		IsFilled: true,
		Conn:     conn,
		Sess:     sess,
	}

	// store session to file
	(&SessionStorage{MgoSession: c.MgoSession}).Save(number, *sess)
}

func (c *BucketSession) IsExist(number string) bool {
	return c.Items[number].IsFilled
}

func (c *BucketSession) Remove(number string) {
	delete(c.Items, number)
	(&SessionStorage{MgoSession: c.MgoSession}).Destroy(number)
}

func (c *BucketSession) Get(number string) *ConnWrapper {
	var wrapperExist bool
	wrapper := c.Items[number]
	if wrapper.IsFilled && wrapper.Sess != nil {
		wrapperExist = true
	}

	session, err := (&SessionStorage{MgoSession: c.MgoSession}).Get(number)

	if err != nil {
		return nil
	}

	if err == nil && !wrapperExist {
		wrapper = ConnWrapper{
			IsFilled: true,
			Conn:     nil,
			Sess:     &session,
		}
		c.Items[number] = wrapper
		return &wrapper
	}

	if err == nil && wrapperExist {
		return &wrapper
	}

	return &wrapper
}

func (c *BucketSession) Reset() {
	sess := SessionStorage{MgoSession: c.MgoSession}
	for number := range c.Items {
		delete(c.Items, number)
		sess.Destroy(number)
	}
}
