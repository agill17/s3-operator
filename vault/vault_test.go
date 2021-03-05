package vault

import (
	"errors"
	vaultApi "github.com/hashicorp/vault/api"
	"testing"
	"time"
)

type mockedVaultLogicalClient struct {
	ReadSecretWithDataResp *vaultApi.Secret
	ReadSecretWithDataErr error
}

func (m mockedVaultLogicalClient) ReadWithData(path string, data map[string][]string) (*vaultApi.Secret, error) {
	return m.ReadSecretWithDataResp, m.ReadSecretWithDataErr
}

func TestClient_ReadSecretPath(t *testing.T) {
	type fields struct {
		vaultClient          *vaultApi.Client
		vaultLogicalClient   vaultLogicalClientInterface
		CreatedLeaseAt       time.Time
		LeaseDuration        int
		ExpectedLeaseToEndAt time.Time
	}
	type args struct {
		path string
		key  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name : "successful read - should return value and no error",
			want: "testValue",
			wantErr: false,
			args: args{
				path: "kv/data/foo/service",
				key:  "testKey",
			},
			fields: fields{
				vaultClient:          nil,
				vaultLogicalClient:   &mockedVaultLogicalClient{
					ReadSecretWithDataResp: &vaultApi.Secret{
						Data: map[string]interface{}{
							"data": map[string]interface{}{
								"testKey": "testValue",
							},
						},
					},
					ReadSecretWithDataErr:  nil,
				},
			},
		},
		{
			name : "key not found - should return empty string and an error",
			want: "",
			wantErr: true,
			args: args{
				path: "kv/data/foo/service",
				key:  "ThisKeyDoesNotExist",
			},
			fields: fields{
				vaultClient:          nil,
				vaultLogicalClient:   &mockedVaultLogicalClient{
					ReadSecretWithDataResp: &vaultApi.Secret{
						Data: map[string]interface{}{
							"data": map[string]interface{}{},
						},
					},
					ReadSecretWithDataErr:  nil,
				},
			},
		},
		{
			name : "Vault went belly up and started crying",
			want: "",
			wantErr: true,
			args: args{
				path: "kv/data/foo/service",
				key:  "testKey",
			},
			fields: fields{
				vaultClient:          nil,
				vaultLogicalClient:   &mockedVaultLogicalClient{
					ReadSecretWithDataResp: nil,
					ReadSecretWithDataErr:  errors.New("ErrVaultWentBellyUp"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Client{
				vaultClient:          tt.fields.vaultClient,
				vaultLogicalClient:   tt.fields.vaultLogicalClient,
				CreatedLeaseAt:       tt.fields.CreatedLeaseAt,
				LeaseDuration:        tt.fields.LeaseDuration,
				ExpectedLeaseToEndAt: tt.fields.ExpectedLeaseToEndAt,
			}
			got, err := v.ReadSecretPath(tt.args.path, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadSecretPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ReadSecretPath() got = %v, want %v", got, tt.want)
			}
		})
	}
}
