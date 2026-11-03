package service_test

import (
	"math"
	"testing"
	"time"

	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestWQMSCalculator_KepmenLH115_And_PermenLHK27(t *testing.T) {
	calc := service.NewWQMSCalculator()

	// 1. Test pH Range Evaluation (6.0 - 9.0, Lavg = 7.5)
	phRes, ok := calc.EvaluateParameter(service.WQMSParameterInput{
		ParamCode:       model.WQMSParamPH,
		SourceType:      "SENSOR",
		RawMeasured:     7.35,
		ValidSamples:    1,
		ExpectedSamples: 1,
	}, model.WaterClass2, 8.0)
	if !ok {
		t.Fatalf("expected pH evaluation to succeed")
	}
	// (7.35 - 7.5) / (6.0 - 7.5) = -0.15 / -1.5 = 0.10
	if math.Abs(phRes.AdjustedRatio-0.10) > 0.01 {
		t.Errorf("expected pH 7.35 ratio ~0.10, got %.4f", phRes.AdjustedRatio)
	}
	if phRes.ExceedsStandard {
		t.Errorf("expected pH 7.35 to satisfy standard")
	}

	// 2. Test DO Inverse Evaluation (Class II >= 4.0 mg/L, Cim = 8.0 mg/L)
	// DO = 5.10 -> (8.0 - 5.10) / (8.0 - 4.0) = 2.90 / 4.0 = 0.725
	doRes, ok := calc.EvaluateParameter(service.WQMSParameterInput{
		ParamCode:       model.WQMSParamDO,
		SourceType:      "SENSOR",
		RawMeasured:     5.10,
		ValidSamples:    1,
		ExpectedSamples: 1,
	}, model.WaterClass2, 8.0)
	if !ok {
		t.Fatalf("expected DO evaluation to succeed")
	}
	if math.Abs(doRes.AdjustedRatio-0.725) > 0.01 {
		t.Errorf("expected DO 5.10 ratio ~0.725, got %.4f", doRes.AdjustedRatio)
	}

	// 3. Test COD Logarithmic Correction (Class II max = 25 mg/L, Measured = 46.2 mg/L)
	// Raw ratio = 46.2 / 25.0 = 1.848 > 1.0 -> Adjusted = 1.0 + 5*log10(1.848) = 1.0 + 5*0.2667 = 2.3335
	codRes, ok := calc.EvaluateParameter(service.WQMSParameterInput{
		ParamCode:       model.WQMSParamCOD,
		SourceType:      "SENSOR",
		RawMeasured:     46.2,
		ValidSamples:    1,
		ExpectedSamples: 1,
	}, model.WaterClass2, 8.0)
	if !ok {
		t.Fatalf("expected COD evaluation to succeed")
	}
	if !codRes.ExceedsStandard || !codRes.LogCorrected {
		t.Errorf("expected COD 46.2 mg/L to exceed Class II 25 mg/L and trigger log correction")
	}
	if math.Abs(codRes.AdjustedRatio-2.3335) > 0.01 {
		t.Errorf("expected COD adjusted ratio ~2.3335, got %.4f", codRes.AdjustedRatio)
	}

	// 4. Test Station IPj Evaluation (Nemerow-Sumitomo)
	stationOut := calc.EvaluateStation([]service.WQMSParameterInput{
		{ParamCode: model.WQMSParamPH, SourceType: "SENSOR", RawMeasured: 7.35, ValidSamples: 24, ExpectedSamples: 24},
		{ParamCode: model.WQMSParamDO, SourceType: "SENSOR", RawMeasured: 5.10, ValidSamples: 24, ExpectedSamples: 24},
		{ParamCode: model.WQMSParamCOD, SourceType: "SENSOR", RawMeasured: 46.2, ValidSamples: 24, ExpectedSamples: 24},
		{ParamCode: model.WQMSParamBOD, SourceType: "SENSOR", RawMeasured: 3.80, ValidSamples: 24, ExpectedSamples: 24},
		{ParamCode: model.WQMSParamTSS, SourceType: "SENSOR", RawMeasured: 38.0, ValidSamples: 24, ExpectedSamples: 24},
		{ParamCode: model.WQMSParamPO4, SourceType: "LAB", RawMeasured: 0.16, ValidSamples: 1, ExpectedSamples: 1},
	}, model.WaterClass2, 8.0)

	if stationOut.OverallCategory != model.WQMSCategoryCemarRingan {
		t.Errorf("expected CEMAR_RINGAN, got %s (IP=%.2f)", stationOut.OverallCategory, stationOut.OverallIP)
	}
	if stationOut.DominantParameter != model.WQMSParamCOD {
		t.Errorf("expected dominant parameter COD, got %s", stationOut.DominantParameter)
	}
	if stationOut.OverallIP <= 1.0 || stationOut.OverallIP > 5.0 {
		t.Errorf("expected 1.0 < IP <= 5.0, got %.2f", stationOut.OverallIP)
	}

	// 5. Test Permen LHK No. 27 Tahun 2021 IKA Weighting (13 Memenuhi * 70 + 17 Cemar Ringan * 50) / 30 = 58.67 -> 58.7 (SEDANG)
	var cats []model.WQMSIPCategory
	for i := 0; i < 13; i++ {
		cats = append(cats, model.WQMSCategoryMemenuhi)
	}
	for i := 0; i < 17; i++ {
		cats = append(cats, model.WQMSCategoryCemarRingan)
	}
	ika := calc.ComputeIKASummary(cats, model.WaterClass2, 30, "REGULATORY_24H")
	if ika.Category != model.IKACategorySedang {
		t.Errorf("expected IKA category SEDANG, got %s (score=%.1f)", ika.Category, ika.Score)
	}
	if math.Abs(ika.Score-58.7) > 0.2 {
		t.Errorf("expected IKA score ~58.7, got %.1f", ika.Score)
	}
}

func TestWQMSService_EndToEndPipeline(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	err = db.AutoMigrate(
		&model.Device{},
		&model.Parameter{},
		&model.AggregationDefinition{},
		&model.AggregationResult{},
		&model.RawData{},
		&model.AuditTrail{},
		&model.WQMSStationConfig{},
		&model.WQMSParameterMapping{},
		&model.WQMSCalculationResult{},
		&model.WQMSParameterResult{},
	)
	if err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	dev := model.Device{
		DeviceCode: "ONLIMO_01",
		DeviceName: "Stasiun ONLIMO Citarum",
		Status:     "ONLINE",
		Enabled:    true,
		Timezone:   "Asia/Jakarta",
	}
	db.Create(&dev)

	pPH := model.Parameter{DeviceID: dev.ID, ParameterCode: "PH", ParameterName: "pH Air", Unit: "pH", Enabled: true}
	pDO := model.Parameter{DeviceID: dev.ID, ParameterCode: "DO", ParameterName: "Dissolved Oxygen", Unit: "mg/L", Enabled: true}
	pCOD := model.Parameter{DeviceID: dev.ID, ParameterCode: "COD", ParameterName: "COD", Unit: "mg/L", Enabled: true}
	db.Create(&pPH)
	db.Create(&pDO)
	db.Create(&pCOD)

	now := time.Now().UTC().Truncate(2 * time.Minute)
	vPH, vDO, vCOD := 7.4, 5.2, 42.0
	for _, item := range []struct {
		pid uint
		val float64
	}{
		{pPH.ID, vPH},
		{pDO.ID, vDO},
		{pCOD.ID, vCOD},
	} {
		v := item.val
		db.Create(&model.AggregationResult{
			AggregationDefinitionID: item.pid,
			DeviceID:                dev.ID,
			ParameterID:             item.pid,
			SourceType:              "PROCESSED",
			PeriodStart:             now.Add(-2 * time.Minute),
			PeriodEnd:               now,
			Identifier:              now.Format("20060102150405"),
			Function:                "AVG",
			IntervalSeconds:         120,
			Value:                   &v,
			AvgValue:                &v,
			ValidCount:              10,
			Quality:                 model.QualityGood,
		})
	}

	wqmsRepo := repository.NewWQMSRepository(db)
	sysRepo := repository.NewSystemRepository(db)
	svc := service.NewWQMSService(db, wqmsRepo, sysRepo)

	// Auto-detects pH, DO, COD mappings and computes IP & IKA on first GetLatestSummary
	bundle, err := svc.GetLatestSummary(t.Context(), 0, "OPERATIONAL")
	if err != nil {
		t.Fatalf("GetLatestSummary failed: %v", err)
	}
	if bundle.IPResult == nil {
		t.Fatalf("expected non-nil IPResult")
	}
	if bundle.IPResult.ActiveParametersCount != 3 {
		t.Errorf("expected 3 auto-mapped parameters, got %d", bundle.IPResult.ActiveParametersCount)
	}
	if bundle.IPResult.DominantParameter != model.WQMSParamCOD {
		t.Errorf("expected dominant COD, got %s", bundle.IPResult.DominantParameter)
	}
	if bundle.IKASummary.Score <= 0 {
		t.Errorf("expected positive IKA score, got %.1f", bundle.IKASummary.Score)
	}

	// Switch to CLASS_4 (COD limit = 80 mg/L -> 42.0 mg/L now satisfies Class IV!)
	updated, err := svc.UpdateStationConfig(t.Context(), model.WQMSStationConfig{
		WaterClass: model.WaterClass4,
	}, "admin", "127.0.0.1")
	if err != nil {
		t.Fatalf("UpdateStationConfig failed: %v", err)
	}
	if updated.IPResult.OverallCategory != model.WQMSCategoryMemenuhi {
		t.Errorf("expected MEMENUHI_BAKU_MUTU under Class IV (COD 42 <= 80), got %s (IP=%.2f)",
			updated.IPResult.OverallCategory, updated.IPResult.OverallIP)
	}
}

