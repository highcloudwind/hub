package wails

import (
	"context"
	"embed"

	"github.com/getAlby/hub/api"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/service"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"gorm.io/gorm"
)

type WailsApp struct {
	ctx context.Context
	svc service.Service
	api api.API
	db  *gorm.DB
}

func NewApp(svc service.Service) *WailsApp {
	return &WailsApp{
		svc: svc,
		api: api.NewAPI(svc, svc.GetDB(), svc.GetConfig(), svc.GetKeys(), svc.GetAlbyOAuthSvc(), svc.GetEventPublisher()),
		db:  svc.GetDB(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (app *WailsApp) startup(ctx context.Context) {
	app.ctx = ctx
}

func LaunchWailsApp(app *WailsApp, assets embed.FS, appIcon []byte) {
	err := wails.Run(&options.App{
		Title:  "AlbyHub",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Logger: NewWailsLogger(),
		// HideWindowOnClose: true, // with this on, there is no way to close the app - wait for v3

		//BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
		Mac: &mac.Options{
			About: &mac.AboutInfo{
				Title: "AlbyHub",
				Icon:  appIcon,
			},
		},
		Linux: &linux.Options{
			Icon: appIcon,
		},
	})

	if err != nil {
		logger.Logger.WithError(err).Error("failed to run Wails app")
	}
}

func NewWailsLogger() WailsLogger {
	return WailsLogger{}
}

type WailsLogger struct {
}

func (wailsLogger WailsLogger) Print(message string) {
	logger.Logger.WithField("wails", true).Print(message)
}

func (wailsLogger WailsLogger) Trace(message string) {
	logger.Logger.WithField("wails", true).Trace(message)
}

func (wailsLogger WailsLogger) Debug(message string) {
	logger.Logger.WithField("wails", true).Debug(message)
}

func (wailsLogger WailsLogger) Info(message string) {
	logger.Logger.WithField("wails", true).Info(message)
}

func (wailsLogger WailsLogger) Warning(message string) {
	logger.Logger.WithField("wails", true).Warning(message)
}

func (wailsLogger WailsLogger) Error(message string) {
	logger.Logger.WithField("wails", true).Error(message)
}

func (wailsLogger WailsLogger) Fatal(message string) {
	logger.Logger.WithField("wails", true).Fatal(message)
}
