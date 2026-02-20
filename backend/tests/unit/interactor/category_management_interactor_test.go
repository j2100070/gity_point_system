package interactor_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/interactor"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// CategoryManagementInteractor テスト
// ========================================

// --- Mock CategoryRepository ---

type mockCategoryRepo struct {
	categories map[uuid.UUID]*entities.Category
	codeMap    map[string]*entities.Category
}

func newMockCategoryRepo() *mockCategoryRepo {
	return &mockCategoryRepo{
		categories: make(map[uuid.UUID]*entities.Category),
		codeMap:    make(map[string]*entities.Category),
	}
}

func (m *mockCategoryRepo) setCategory(c *entities.Category) {
	m.categories[c.ID] = c
	m.codeMap[c.Code] = c
}

func (m *mockCategoryRepo) Create(ctx context.Context, category *entities.Category) error {
	m.categories[category.ID] = category
	m.codeMap[category.Code] = category
	return nil
}
func (m *mockCategoryRepo) Read(ctx context.Context, id uuid.UUID) (*entities.Category, error) {
	c, ok := m.categories[id]
	if !ok {
		return nil, errors.New("category not found")
	}
	return c, nil
}
func (m *mockCategoryRepo) ReadByCode(ctx context.Context, code string) (*entities.Category, error) {
	c, ok := m.codeMap[code]
	if !ok {
		return nil, errors.New("category not found")
	}
	return c, nil
}
func (m *mockCategoryRepo) Update(ctx context.Context, category *entities.Category) error {
	m.categories[category.ID] = category
	return nil
}
func (m *mockCategoryRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.categories, id)
	return nil
}
func (m *mockCategoryRepo) ReadList(ctx context.Context, activeOnly bool) ([]*entities.Category, error) {
	result := make([]*entities.Category, 0)
	for _, c := range m.categories {
		if activeOnly && !c.IsActive {
			continue
		}
		result = append(result, c)
	}
	return result, nil
}
func (m *mockCategoryRepo) Count(ctx context.Context) (int64, error) {
	return int64(len(m.categories)), nil
}
func (m *mockCategoryRepo) ExistsCode(ctx context.Context, code string, excludeID *uuid.UUID) (bool, error) {
	c, ok := m.codeMap[code]
	if !ok {
		return false, nil
	}
	if excludeID != nil && c.ID == *excludeID {
		return false, nil
	}
	return true, nil
}

// --- CreateCategory ---

func TestCategoryManagementInteractor_CreateCategory(t *testing.T) {
	setup := func() (*mockCategoryRepo, inputport.CategoryManagementInputPort) {
		catRepo := newMockCategoryRepo()
		sut := interactor.NewCategoryManagementInteractor(catRepo, &mockLogger{})
		return catRepo, sut
	}

	t.Run("正常にカテゴリを作成できる", func(t *testing.T) {
		_, sut := setup()

		resp, err := sut.CreateCategory(context.Background(), &inputport.CreateCategoryRequest{
			Name: "飲み物", Code: "drink", Description: "飲み物カテゴリ", DisplayOrder: 1,
		})
		require.NoError(t, err)
		assert.NotNil(t, resp.Category)
		assert.Equal(t, "飲み物", resp.Category.Name)
		assert.Equal(t, "drink", resp.Category.Code)
	})

	t.Run("コードが重複している場合エラー", func(t *testing.T) {
		catRepo, sut := setup()
		existing, _ := entities.NewCategory("既存", "drink", "既存カテゴリ", 1)
		catRepo.setCategory(existing)

		_, err := sut.CreateCategory(context.Background(), &inputport.CreateCategoryRequest{
			Name: "飲み物", Code: "drink", Description: "飲み物カテゴリ", DisplayOrder: 2,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("名前が空の場合エラー", func(t *testing.T) {
		_, sut := setup()

		_, err := sut.CreateCategory(context.Background(), &inputport.CreateCategoryRequest{
			Name: "", Code: "empty", Description: "", DisplayOrder: 1,
		})
		assert.Error(t, err)
	})
}

// --- UpdateCategory ---

func TestCategoryManagementInteractor_UpdateCategory(t *testing.T) {
	setup := func() (*mockCategoryRepo, inputport.CategoryManagementInputPort) {
		catRepo := newMockCategoryRepo()
		sut := interactor.NewCategoryManagementInteractor(catRepo, &mockLogger{})
		return catRepo, sut
	}

	t.Run("正常にカテゴリを更新できる", func(t *testing.T) {
		catRepo, sut := setup()
		cat, _ := entities.NewCategory("飲み物", "drink", "古い説明", 1)
		catRepo.setCategory(cat)

		resp, err := sut.UpdateCategory(context.Background(), &inputport.UpdateCategoryRequest{
			CategoryID: cat.ID, Name: "ドリンク", Description: "新しい説明", DisplayOrder: 2, IsActive: true,
		})
		require.NoError(t, err)
		assert.Equal(t, "ドリンク", resp.Category.Name)
		assert.Equal(t, "新しい説明", resp.Category.Description)
	})

	t.Run("存在しないカテゴリの場合エラー", func(t *testing.T) {
		_, sut := setup()

		_, err := sut.UpdateCategory(context.Background(), &inputport.UpdateCategoryRequest{
			CategoryID: uuid.New(), Name: "test", Description: "", DisplayOrder: 1, IsActive: true,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "category not found")
	})
}

// --- DeleteCategory ---

func TestCategoryManagementInteractor_DeleteCategory(t *testing.T) {
	setup := func() (*mockCategoryRepo, inputport.CategoryManagementInputPort) {
		catRepo := newMockCategoryRepo()
		sut := interactor.NewCategoryManagementInteractor(catRepo, &mockLogger{})
		return catRepo, sut
	}

	t.Run("正常にカテゴリを削除できる", func(t *testing.T) {
		catRepo, sut := setup()
		cat, _ := entities.NewCategory("削除対象", "delete_me", "説明", 1)
		catRepo.setCategory(cat)

		err := sut.DeleteCategory(context.Background(), &inputport.DeleteCategoryRequest{
			CategoryID: cat.ID,
		})
		assert.NoError(t, err)
	})

	t.Run("存在しないカテゴリの場合エラー", func(t *testing.T) {
		_, sut := setup()

		err := sut.DeleteCategory(context.Background(), &inputport.DeleteCategoryRequest{
			CategoryID: uuid.New(),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "category not found")
	})
}

// --- GetCategoryList ---

func TestCategoryManagementInteractor_GetCategoryList(t *testing.T) {
	t.Run("正常にカテゴリ一覧を取得できる", func(t *testing.T) {
		catRepo := newMockCategoryRepo()
		sut := interactor.NewCategoryManagementInteractor(catRepo, &mockLogger{})

		cat1, _ := entities.NewCategory("飲み物", "drink", "", 1)
		cat2, _ := entities.NewCategory("お菓子", "snack", "", 2)
		catRepo.setCategory(cat1)
		catRepo.setCategory(cat2)

		resp, err := sut.GetCategoryList(context.Background(), &inputport.GetCategoryListRequest{
			ActiveOnly: false,
		})
		require.NoError(t, err)
		assert.Equal(t, 2, len(resp.Categories))
		assert.Equal(t, int64(2), resp.Total)
	})
}
