package service_test

import (
	"context"
	"testing"
	"time"

	"datalogger/internal/database"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestAlarmSync_ClearsOrphanAlarmsOnDeleteOrDisable(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	ctx := context.Background()
	deviceRepo := repository.NewDeviceRepository(db)
	systemRepo := repository.NewSystemRepository(db)
	alarmRepo := repository.NewAlarmRepository(db)

	alarmService := service.NewAlarmService(alarmRepo, systemRepo)
	deviceService := service.NewDeviceService(deviceRepo, systemRepo)
	deviceService.SetAlarmService(alarmService)

	evaluator := service.NewAlarmEvaluator(
		ctx,
		alarmService,
		service.DefaultAlarmEvaluatorConfig(),
		func(ctx context.Context, paramID uint) (*model.Parameter, error) {
			return deviceRepo.FindParameterByID(paramID)
		},
	)
	evaluator.SetDeviceLookup(func(ctx context.Context, deviceID uint) (*model.Device, error) {
		return deviceRepo.GetByID(deviceID)
	})
	alarmService.SetAlarmEvaluator(evaluator)

	// 1. Create active device + 2 parameters
	dev := &model.Device{
		DeviceCode:       "com2",
		DeviceName:       "COM2 Analyzer",
		DeviceType:       "MODBUS_RTU",
		Status:           model.DeviceStatusActive,
		Enabled:          true,
		ConnectionStatus: model.DeviceConnOnline,
	}
	if err := deviceRepo.Create(dev); err != nil {
		t.Fatalf("failed to create device: %v", err)
	}

	warnHigh := 100.0
	p1 := &model.Parameter{
		DeviceID:      dev.ID,
		ParameterCode: "pm25",
		ParameterName: "PM 2.5",
		DataType:      model.DataTypeFloat32,
		Enabled:       true,
		WarningHigh:   &warnHigh,
	}
	p2 := &model.Parameter{
		DeviceID:      dev.ID,
		ParameterCode: "pm10",
		ParameterName: "PM 10",
		DataType:      model.DataTypeFloat32,
		Enabled:       true,
		WarningHigh:   &warnHigh,
	}
	if err := deviceRepo.CreateParameter(p1); err != nil {
		t.Fatalf("failed to create p1: %v", err)
	}
	if err := deviceRepo.CreateParameter(p2); err != nil {
		t.Fatalf("failed to create p2: %v", err)
	}

	now := time.Now().UTC()
	evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
		DeviceID:       dev.ID,
		ParameterID:    p1.ID,
		ProcessedValue: 150.0,
		Quality:        model.QualityGood,
		ReceivedAt:     now,
	}, p1)
	evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
		DeviceID:       dev.ID,
		ParameterID:    p2.ID,
		ProcessedValue: 180.0,
		Quality:        model.QualityGood,
		ReceivedAt:     now,
	}, p2)

	active, err := alarmService.GetAlarmsByStatus(ctx, "ACTIVE")
	if err != nil {
		t.Fatalf("GetAlarmsByStatus failed: %v", err)
	}
	if len(active) != 2 || evaluator.GetActiveCount() != 2 {
		t.Fatalf("expected 2 active alarms in DB and evaluator, got db=%d eval=%d", len(active), evaluator.GetActiveCount())
	}

	// 2. Soft-delete parameter p1 -> its alarm must immediately clear in DB and evaluator
	if err := deviceService.DeleteParameter(dev.ID, p1.ID, "admin", "127.0.0.1", "test"); err != nil {
		t.Fatalf("DeleteParameter failed: %v", err)
	}
	active, _ = alarmService.GetAlarmsByStatus(ctx, "ACTIVE")
	if len(active) != 1 || evaluator.GetActiveCount() != 1 {
		t.Fatalf("after deleting p1, expected 1 active alarm in DB and evaluator, got db=%d eval=%d", len(active), evaluator.GetActiveCount())
	}

	// 3. Soft-delete device com2 -> remaining alarm on p2 must immediately clear in DB and evaluator
	if err := deviceService.DeleteDevice(dev.ID, "admin", "127.0.0.1", "test"); err != nil {
		t.Fatalf("DeleteDevice failed: %v", err)
	}
	active, _ = alarmService.GetAlarmsByStatus(ctx, "ACTIVE")
	if len(active) != 0 || evaluator.GetActiveCount() != 0 {
		t.Fatalf("after deleting device, expected 0 active alarms in DB and evaluator, got db=%d eval=%d", len(active), evaluator.GetActiveCount())
	}
}

