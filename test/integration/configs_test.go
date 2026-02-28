package integration

import (
	"context"
	"testing"

	mgmtConfig "github.com/openziti/edge-api/rest_management_api_client/config"
	"github.com/openziti/edge-api/rest_model"
)

func TestCreateAndListConfigType(t *testing.T) {
	ctx := context.Background()
	name := "test-ct-" + uniqueSuffix()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	createResp, err := mgmt.Config.CreateConfigType(
		mgmtConfig.NewCreateConfigTypeParams().WithContext(ctx).WithConfigType(&rest_model.ConfigTypeCreate{
			Name: &name,
		}), nil)
	if err != nil {
		t.Fatalf("create config type: %v", err)
	}
	ctID := createResp.GetPayload().Data.ID
	defer deleteConfigType(t, ctx, ctID)

	listResp, err := mgmt.Config.ListConfigTypes(
		mgmtConfig.NewListConfigTypesParams().WithContext(ctx).WithFilter(ptr(`name = "`+name+`"`)), nil)
	if err != nil {
		t.Fatalf("list config types: %v", err)
	}
	if len(listResp.GetPayload().Data) == 0 {
		t.Fatalf("expected config type %q in list, got 0 results", name)
	}
}

func TestGetConfigType(t *testing.T) {
	ctx := context.Background()
	name := "test-ct-get-" + uniqueSuffix()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	createResp, err := mgmt.Config.CreateConfigType(
		mgmtConfig.NewCreateConfigTypeParams().WithContext(ctx).WithConfigType(&rest_model.ConfigTypeCreate{
			Name: &name,
		}), nil)
	if err != nil {
		t.Fatalf("create config type: %v", err)
	}
	ctID := createResp.GetPayload().Data.ID
	defer deleteConfigType(t, ctx, ctID)

	getResp, err := mgmt.Config.DetailConfigType(
		mgmtConfig.NewDetailConfigTypeParams().WithContext(ctx).WithID(ctID), nil)
	if err != nil {
		t.Fatalf("get config type: %v", err)
	}
	if *getResp.GetPayload().Data.ID != ctID {
		t.Errorf("expected id %q, got %q", ctID, getResp.GetPayload().Data.ID)
	}
}

func TestCreateAndListConfig(t *testing.T) {
	ctx := context.Background()
	ctName := "test-ct-for-cfg-" + uniqueSuffix()
	cfgName := "test-cfg-" + uniqueSuffix()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	ctResp, err := mgmt.Config.CreateConfigType(
		mgmtConfig.NewCreateConfigTypeParams().WithContext(ctx).WithConfigType(&rest_model.ConfigTypeCreate{
			Name: &ctName,
		}), nil)
	if err != nil {
		t.Fatalf("create config type: %v", err)
	}
	ctID := ctResp.GetPayload().Data.ID
	defer deleteConfigType(t, ctx, ctID)

	data := map[string]any{"host": "localhost", "port": 8080}
	createResp, err := mgmt.Config.CreateConfig(
		mgmtConfig.NewCreateConfigParams().WithContext(ctx).WithConfig(&rest_model.ConfigCreate{
			Name:         &cfgName,
			ConfigTypeID: &ctID,
			Data:         &data,
		}), nil)
	if err != nil {
		t.Fatalf("create config: %v", err)
	}
	cfgID := createResp.GetPayload().Data.ID
	defer deleteConfig(t, ctx, cfgID)

	listResp, err := mgmt.Config.ListConfigs(
		mgmtConfig.NewListConfigsParams().WithContext(ctx).WithFilter(ptr(`name = "`+cfgName+`"`)), nil)
	if err != nil {
		t.Fatalf("list configs: %v", err)
	}
	if len(listResp.GetPayload().Data) == 0 {
		t.Fatalf("expected config %q in list, got 0 results", cfgName)
	}
}

func TestGetConfig(t *testing.T) {
	ctx := context.Background()
	ctName := "test-ct-get-cfg-" + uniqueSuffix()
	cfgName := "test-cfg-get-" + uniqueSuffix()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	ctResp, err := mgmt.Config.CreateConfigType(
		mgmtConfig.NewCreateConfigTypeParams().WithContext(ctx).WithConfigType(&rest_model.ConfigTypeCreate{Name: &ctName}), nil)
	if err != nil {
		t.Fatalf("create config type: %v", err)
	}
	ctID := ctResp.GetPayload().Data.ID
	defer deleteConfigType(t, ctx, ctID)

	data := any(map[string]any{"key": "value"})
	cfgResp, err := mgmt.Config.CreateConfig(
		mgmtConfig.NewCreateConfigParams().WithContext(ctx).WithConfig(&rest_model.ConfigCreate{
			Name:         &cfgName,
			ConfigTypeID: &ctID,
			Data:         &data,
		}), nil)
	if err != nil {
		t.Fatalf("create config: %v", err)
	}
	cfgID := cfgResp.GetPayload().Data.ID
	defer deleteConfig(t, ctx, cfgID)

	getResp, err := mgmt.Config.DetailConfig(
		mgmtConfig.NewDetailConfigParams().WithContext(ctx).WithID(cfgID), nil)
	if err != nil {
		t.Fatalf("get config: %v", err)
	}
	if *getResp.GetPayload().Data.ID != cfgID {
		t.Errorf("expected id %q, got %q", cfgID, getResp.GetPayload().Data.ID)
	}
}

func TestUpdateConfig(t *testing.T) {
	ctx := context.Background()
	ctName := "test-ct-upd-cfg-" + uniqueSuffix()
	cfgName := "test-cfg-update-" + uniqueSuffix()
	updatedName := cfgName + "-updated"

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	ctResp, err := mgmt.Config.CreateConfigType(
		mgmtConfig.NewCreateConfigTypeParams().WithContext(ctx).WithConfigType(&rest_model.ConfigTypeCreate{Name: &ctName}), nil)
	if err != nil {
		t.Fatalf("create config type: %v", err)
	}
	ctID := ctResp.GetPayload().Data.ID
	defer deleteConfigType(t, ctx, ctID)

	data := any(map[string]any{"v": 1})
	cfgResp, err := mgmt.Config.CreateConfig(
		mgmtConfig.NewCreateConfigParams().WithContext(ctx).WithConfig(&rest_model.ConfigCreate{
			Name:         &cfgName,
			ConfigTypeID: &ctID,
			Data:         &data,
		}), nil)
	if err != nil {
		t.Fatalf("create config: %v", err)
	}
	cfgID := cfgResp.GetPayload().Data.ID
	defer deleteConfig(t, ctx, cfgID)

	newData := any(map[string]any{"v": 2})
	_, err = mgmt.Config.UpdateConfig(
		mgmtConfig.NewUpdateConfigParams().WithContext(ctx).WithID(cfgID).WithConfig(&rest_model.ConfigUpdate{
			Name: &updatedName,
			Data: &newData,
		}), nil)
	if err != nil {
		t.Fatalf("update config: %v", err)
	}

	getResp, err := mgmt.Config.DetailConfig(
		mgmtConfig.NewDetailConfigParams().WithContext(ctx).WithID(cfgID), nil)
	if err != nil {
		t.Fatalf("get config after update: %v", err)
	}
	if *getResp.GetPayload().Data.Name != updatedName {
		t.Errorf("expected name %q, got %q", updatedName, *getResp.GetPayload().Data.Name)
	}
}

func TestDeleteConfig(t *testing.T) {
	ctx := context.Background()
	ctName := "test-ct-del-cfg-" + uniqueSuffix()
	cfgName := "test-cfg-delete-" + uniqueSuffix()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	ctResp, err := mgmt.Config.CreateConfigType(
		mgmtConfig.NewCreateConfigTypeParams().WithContext(ctx).WithConfigType(&rest_model.ConfigTypeCreate{Name: &ctName}), nil)
	if err != nil {
		t.Fatalf("create config type: %v", err)
	}
	ctID := ctResp.GetPayload().Data.ID
	defer deleteConfigType(t, ctx, ctID)

	data := any(map[string]any{})
	cfgResp, err := mgmt.Config.CreateConfig(
		mgmtConfig.NewCreateConfigParams().WithContext(ctx).WithConfig(&rest_model.ConfigCreate{
			Name:         &cfgName,
			ConfigTypeID: &ctID,
			Data:         &data,
		}), nil)
	if err != nil {
		t.Fatalf("create config: %v", err)
	}
	cfgID := cfgResp.GetPayload().Data.ID

	if _, err := mgmt.Config.DeleteConfig(
		mgmtConfig.NewDeleteConfigParams().WithContext(ctx).WithID(cfgID), nil); err != nil {
		t.Fatalf("delete config: %v", err)
	}

	listResp, err := mgmt.Config.ListConfigs(
		mgmtConfig.NewListConfigsParams().WithContext(ctx).WithFilter(ptr(`name = "`+cfgName+`"`)), nil)
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(listResp.GetPayload().Data) > 0 {
		t.Error("expected config to be deleted, but it still appears in list")
	}
}

func deleteConfigType(t *testing.T, ctx context.Context, id string) {
	t.Helper()
	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Logf("WARN: cleanup config type %q: get client: %v", id, err)
		return
	}
	if _, err := mgmt.Config.DeleteConfigType(
		mgmtConfig.NewDeleteConfigTypeParams().WithContext(ctx).WithID(id), nil); err != nil {
		t.Logf("WARN: cleanup config type %q: %v", id, err)
	}
}

func deleteConfig(t *testing.T, ctx context.Context, id string) {
	t.Helper()
	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Logf("WARN: cleanup config %q: get client: %v", id, err)
		return
	}
	if _, err := mgmt.Config.DeleteConfig(
		mgmtConfig.NewDeleteConfigParams().WithContext(ctx).WithID(id), nil); err != nil {
		t.Logf("WARN: cleanup config %q: %v", id, err)
	}
}
