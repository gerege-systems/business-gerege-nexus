/*
 * Gerege Nexus — Commerce
 * Copyright (c) 2026 Gerege Systems Development Team, Gerege Nomadica Foundation.
 * Distributed under the Apache 2.0 License.
 */

// The contact register: customers, suppliers and the organisations this one
// deals with.
//
// It came here from the platform, and it went two ways to get here. It was an
// app of its own; it was absorbed into the directory on 2026-08-15, on the
// argument that "who this organisation is made of" and "who it deals with" are
// one subject; and it left again the same day, because the second half of that
// subject is not the platform's. Everybody has departments and staff. Only a
// business that sells has customers and suppliers, and that is what a
// distribution is for.
//
// What stayed behind: the `contacts` table, created by the platform's migration
// 00003 and applied on every deployment in the field — an applied migration is
// not removed for tidiness — and the screens, because the shell is one image
// serving every deployment and carries the union of first-party pages. Without
// this module behind them those pages are inert: unlisted in the menu, which is
// built from registered modules, and refused by the API.
package contacts

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gerege-systems/open-gerege-nexus/backend/pkg/nexus"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Contact struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Company   string    `json:"company"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Module struct {
	db nexus.DB
}

func New(p nexus.Platform) *Module {
	m := &Module{db: p.DB()}
	nexus.Register(m)
	return m
}

func (m *Module) ID() string { return "io.gerege.nexus.contacts" }

// MenuPermission and RoutePermissionPrefix are this module's half of
// nexus.AccessPolicy: the platform gates /api/v1/contacts with contacts.read on
// a GET and contacts.manage on anything else, and hides the menu entry from
// somebody who holds neither.
//
// Stated here rather than in the platform, which is what lets the module live
// in this repository at all — the platform used to hold a switch keyed by app
// id, and a switch cannot follow an app to another product.
func (m *Module) MenuPermission() string        { return "contacts.read" }
func (m *Module) RoutePermissionPrefix() string { return "contacts" }
func (m *Module) Name() string                  { return "Contacts" }

// 2.0.0: the app is the same code and not the same app. It was folded into the
// directory and taken out again, which means an installation of 1.0.0 is an
// installation of something that belonged to another product.
func (m *Module) Version() string { return "2.0.0" }

func (m *Module) Dependencies() []nexus.Dependency {
	return nil
}

func (m *Module) Permissions() []nexus.PermissionDefinition {
	return []nexus.PermissionDefinition{
		{Code: "contacts.read", Name: "Read Contacts", Description: "View contacts list"},
		{Code: "contacts.manage", Name: "Manage Contacts", Description: "Create and edit contacts"},
	}
}

// Menus are declared here with their paths rather than left to a blueprint.
//
// A blueprint lives in the platform's menu package, keyed by app id, and a
// distribution cannot add to it — which is right: the platform should not be
// carrying a list of screens for products it does not ship. So the three that
// used to be blueprint entries are ordinary menu definitions, pointing at the
// pages the shell already carries.
func (m *Module) Menus() []nexus.MenuDefinition {
	return []nexus.MenuDefinition{
		{ID: "contacts", Label: "Contacts", Path: "/contacts", Icon: "users", Order: 10, Labels: map[string]string{
			"mn": "Харилцагчид", "ar": "جهات الاتصال", "zh": "联系人", "fr": "Contacts", "ru": "Контакты", "es": "Contactos"}},
		{ID: "contacts_segments", Label: "Segments", Path: "/module/contacts/segments", Icon: "users", Order: 20, Labels: map[string]string{
			"mn": "Сегментүүд", "ar": "الشرائح", "zh": "客户分组", "fr": "Segments", "ru": "Сегменты", "es": "Segmentos"}},
		{ID: "contacts_duplicates", Label: "Duplicates", Path: "/module/contacts/duplicates", Icon: "copy", Order: 30, Labels: map[string]string{
			"mn": "Давхардал", "ar": "التكرارات", "zh": "重复记录", "fr": "Doublons", "ru": "Дубликаты", "es": "Duplicados"}},
		{ID: "contacts_import", Label: "Import contacts", Path: "/module/contacts/import", Icon: "upload", Order: 40, Labels: map[string]string{
			"mn": "Импорт", "ar": "استيراد جهات الاتصال", "zh": "导入联系人", "fr": "Importer des contacts", "ru": "Импорт контактов", "es": "Importar contactos"}},
	}
}

func (m *Module) RegisterRoutes(r chi.Router, tenantAuthMiddleware func(http.Handler) http.Handler) {
	r.Route("/api/v1/contacts", func(cr chi.Router) {
		cr.Use(tenantAuthMiddleware)
		cr.Get("/", m.listContactsHandler)
		cr.Post("/", m.createContactHandler)
		cr.Put("/{id}", m.updateContactHandler)
	})
}

func (m *Module) listContactsHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := nexus.RequireTenant(w, r)
	if !ok {
		return
	}

	rows, err := m.db.Query(r.Context(),
		`SELECT id, tenant_id, name, email, phone, company, active, created_at, updated_at 
		 FROM contacts WHERE tenant_id = $1 ORDER BY name ASC`, tenantID)
	if err != nil {
		nexus.Error(w, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()

	list := make([]Contact, 0)
	for rows.Next() {
		var c Contact
		if err := rows.Scan(&c.ID, &c.TenantID, &c.Name, &c.Email, &c.Phone, &c.Company, &c.Active, &c.CreatedAt, &c.UpdatedAt); err != nil {
			nexus.Error(w, http.StatusInternalServerError, "scan error")
			return
		}
		list = append(list, c)
	}
	// A stream that breaks partway through ends the loop the same way a
	// complete one does, so without this the caller receives a short list
	// under a 200 and has no way to tell it apart from the whole set.
	if err := rows.Err(); err != nil {
		nexus.Error(w, http.StatusInternalServerError, "scan error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list)
}

func (m *Module) createContactHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := nexus.UserFromContext(r.Context())
	if err != nil {
		nexus.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Phone   string `json:"phone"`
		Company string `json:"company"`
		Active  bool   `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		nexus.Error(w, http.StatusBadRequest, "invalid contact payload, name is required")
		return
	}

	id := uuid.New().String()
	now := time.Now()

	_, err = m.db.Exec(r.Context(),
		`INSERT INTO contacts (id, tenant_id, name, email, phone, company, active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		id, claims.TenantID, req.Name, req.Email, req.Phone, req.Company, req.Active, now, now)
	if err != nil {
		nexus.Error(w, http.StatusInternalServerError, "failed to create contact")
		return
	}

	contact := Contact{
		ID:        id,
		TenantID:  claims.TenantID,
		Name:      req.Name,
		Email:     req.Email,
		Phone:     req.Phone,
		Company:   req.Company,
		Active:    req.Active,
		CreatedAt: now,
		UpdatedAt: now,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(contact)
}

func (m *Module) updateContactHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := nexus.UserFromContext(r.Context())
	if err != nil {
		nexus.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := chi.URLParam(r, "id")
	var req struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Phone   string `json:"phone"`
		Company string `json:"company"`
		Active  bool   `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		nexus.Error(w, http.StatusBadRequest, "invalid payload")
		return
	}

	now := time.Now()
	res, err := m.db.Exec(r.Context(),
		`UPDATE contacts SET name = $1, email = $2, phone = $3, company = $4, active = $5, updated_at = $6
		 WHERE id = $7 AND tenant_id = $8`,
		req.Name, req.Email, req.Phone, req.Company, req.Active, now, id, claims.TenantID)
	if err != nil || res.RowsAffected() == 0 {
		nexus.Error(w, http.StatusNotFound, "contact not found or unauthorized")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}
