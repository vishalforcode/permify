package postgres

import (
	"context"
	"os"

	"github.com/rs/xid"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/database"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

var _ = Describe("SchemaReader", func() {
	var db database.Database
	var schemaWriter *SchemaWriter
	var schemaReader *SchemaReader

	BeforeEach(func() {
		version := os.Getenv("POSTGRES_VERSION")

		if version == "" {
			version = "14"
		}

		db = postgresDB(version)
		schemaWriter = NewSchemaWriter(db.(*PQDatabase.Postgres))
		schemaReader = NewSchemaReader(db.(*PQDatabase.Postgres))
	})

	AfterEach(func() {
		err := db.Close()
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("Write Schema", func() {
		It("should write schema for a tenant", func() {
			ctx := context.Background()

			version := xid.New().String()

			schema := []storage.SchemaDefinition{
				{TenantID: "t1", Name: "user", SerializedDefinition: []byte("entity user {}"), Version: version},
				{TenantID: "t1", Name: "organization", SerializedDefinition: []byte("entity organization { relation admin @user}"), Version: version},
			}

			err := schemaWriter.WriteSchema(ctx, schema)
			Expect(err).ShouldNot(HaveOccurred())

			sch, err := schemaReader.ReadSchema(ctx, "t1", version)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(sch.EntityDefinitions["user"]).Should(Equal(&base.EntityDefinition{
				Name:        "user",
				Relations:   map[string]*base.RelationDefinition{},
				Permissions: map[string]*base.PermissionDefinition{},
				Attributes:  map[string]*base.AttributeDefinition{},
				References:  map[string]base.EntityDefinition_Reference{},
			}))

			Expect(sch.EntityDefinitions["organization"]).Should(Equal(&base.EntityDefinition{
				Name: "organization",
				Relations: map[string]*base.RelationDefinition{
					"admin": {
						Name: "admin",
						RelationReferences: []*base.RelationReference{
							{
								Type:     "user",
								Relation: "",
							},
						},
					},
				},
				Permissions: map[string]*base.PermissionDefinition{},
				Attributes:  map[string]*base.AttributeDefinition{},
				References: map[string]base.EntityDefinition_Reference{
					"admin": base.EntityDefinition_REFERENCE_RELATION,
				},
			},
			))

			Expect(sch.RuleDefinitions).Should(Equal(map[string]*base.RuleDefinition{}))

			Expect(sch.References["user"]).Should(Equal(base.SchemaDefinition_REFERENCE_ENTITY))
			Expect(sch.References["organization"]).Should(Equal(base.SchemaDefinition_REFERENCE_ENTITY))
		})
	})
})
