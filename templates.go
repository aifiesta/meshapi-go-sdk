package meshapi

import "context"

// TemplatesResource provides access to the /v1/templates CRUD endpoints.
type TemplatesResource struct {
	http *httpClient
}

// Create creates a new prompt template.
func (r *TemplatesResource) Create(ctx context.Context, params CreateTemplateParams) (*TemplateSummary, error) {
	var out TemplateSummary
	if err := r.http.post(ctx, "/v1/templates", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns all templates owned by the authenticated user.
func (r *TemplatesResource) List(ctx context.Context) ([]TemplateSummary, error) {
	var out []TemplateSummary
	if err := r.http.get(ctx, "/v1/templates", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Get returns a single template by ID.
func (r *TemplatesResource) Get(ctx context.Context, id string) (*TemplateSummary, error) {
	var out TemplateSummary
	if err := r.http.get(ctx, "/v1/templates/"+id, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update partially updates a template.
func (r *TemplatesResource) Update(ctx context.Context, id string, params UpdateTemplateParams) (*TemplateSummary, error) {
	var out TemplateSummary
	if err := r.http.patch(ctx, "/v1/templates/"+id, params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete deletes a template (returns nil on 204 No Content).
func (r *TemplatesResource) Delete(ctx context.Context, id string) error {
	return r.http.delete(ctx, "/v1/templates/"+id)
}
