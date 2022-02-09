package main

import (
	"net/http"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/matrix-org/gomatrix"
)

const perfMatrixKey = "auth.matrix"
const prefMatrixSerKey = perfMatrixKey + ".server"
const prefMatrixUserKey = perfMatrixKey + ".username"
const prefMatrixAuthToken = perfMatrixKey + ".token"

type matrix struct {
	app    fyne.App
	client *gomatrix.Client
}

func initMatrix(a fyne.App) service {
	return &matrix{app: a}
}

func (t *matrix) configure(u *ui) (fyne.CanvasObject, func(prefix string, a fyne.App)) {
	matSer := widget.NewEntry()
	matUser := widget.NewEntry()
	matPass := widget.NewPasswordEntry()
	return widget.NewForm(
			&widget.FormItem{Text: "Matrix server", Widget: matSer},
			&widget.FormItem{Text: "Matrix username", Widget: matUser},
			&widget.FormItem{Text: "Matrix password", Widget: matPass}),
		func(prefix string, a fyne.App) {
			a.Preferences().SetString(prefix+prefMatrixSerKey, matSer.Text)
			a.Preferences().SetString(prefix+prefMatrixUserKey, matUser.Text)

			var err error
			t.client, err = gomatrix.NewClient(
				matSer.Text,
				"",
				"",
			)
			resp, err := t.client.Login(&gomatrix.ReqLogin{
				Type:     "m.login.password",
				User:     matUser.Text,
				Password: matPass.Text,
			})
			if err != nil {
				fyne.LogError("Login failed", err)
				return
			}
			a.Preferences().SetString(prefix+prefMatrixAuthToken, resp.AccessToken)
			t.login(prefix, u)
		}
}

func (t *matrix) disconnect() {
}

func (t *matrix) login(prefix string, u *ui) {
	var err error
	t.client, err = gomatrix.NewClient(
		t.app.Preferences().String(prefix+prefMatrixSerKey),
		"",
		"",
	)
	if err != nil {
		fyne.LogError("Client creation failed", err)
		return
	}
	username := t.app.Preferences().String(prefix + prefMatrixUserKey)
	t.client.SetCredentials(username,
		t.app.Preferences().String(prefix+prefMatrixAuthToken))
	syncer := gomatrix.NewDefaultSyncer(username, gomatrix.NewInMemoryStore())
	t.client.Client = http.DefaultClient
	t.client.Syncer = syncer

	syncer.OnEventType("m.room.message", func(ev *gomatrix.Event) {
		if ev.Sender == username {
			return
		}
	})
	go func() {
		if err := t.client.Sync(); err != nil {
			fyne.LogError("Sync failed", err)
			return
		}
	}()

}

func (t *matrix) send(ch *channel, text string) {
}
