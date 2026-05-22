package settings

import "testing"

func TestMigrateActivityOffset(t *testing.T) {
	app := App{Version: 0, ActivityPanelOffset: 0}
	migrate(&app)
	if app.Version != fileVersion {
		t.Fatalf("version %d", app.Version)
	}
	if app.ActivityPanelOffset != 0.4 {
		t.Fatalf("offset %v", app.ActivityPanelOffset)
	}
	if app.UIScale != 0 {
		t.Fatalf("ui scale %v", app.UIScale)
	}
}

func TestMigrateLegacyUIScaleToAuto(t *testing.T) {
	app := App{Version: 3, UIScale: 1.15}
	migrate(&app)
	if app.Version != fileVersion {
		t.Fatalf("version %d", app.Version)
	}
	if app.UIScale != 0 {
		t.Fatalf("ui scale %v want auto (0)", app.UIScale)
	}
}

func TestValidateUIScaleAuto(t *testing.T) {
	app := Default()
	if err := Validate(app); err != nil {
		t.Fatal(err)
	}
}

func TestValidateQueueWorkers(t *testing.T) {
	app := Default()
	app.QueueParallel = 99
	if err := Validate(app); err == nil {
		t.Fatal("expected error")
	}
}
