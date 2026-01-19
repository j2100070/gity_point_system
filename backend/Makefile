.PHONY: test test-unit test-integration test-e2e mock clean

MOCKGEN := $(shell go env GOPATH)/bin/mockgen

# モック生成
mock:
	@echo "Generating mocks..."
	$(MOCKGEN) -source=internal/domain/user.go -destination=internal/domain/mock/mock_user_repository.go -package=mock
	$(MOCKGEN) -source=internal/domain/transaction.go -destination=internal/domain/mock/mock_transaction_repository.go -package=mock
	$(MOCKGEN) -source=internal/domain/qrcode.go -destination=internal/domain/mock/mock_qrcode_repository.go -package=mock
	$(MOCKGEN) -source=internal/domain/friendship.go -destination=internal/domain/mock/mock_friendship_repository.go -package=mock
	$(MOCKGEN) -source=internal/domain/session.go -destination=internal/domain/mock/mock_session_repository.go -package=mock
	@echo "Mocks generated successfully!"

# 単体テスト
test-unit:
	@echo "Running unit tests..."
	go test -v -race -coverprofile=coverage-unit.out ./internal/domain/... ./internal/usecase/...

# 結合テスト（実際のDB使用、ポイントの実際の値を検証）
test-integration:
	@echo "Running integration tests..."
	@echo "Note: Requires PostgreSQL test database 'gity_point_test'"
	go test -v -race -tags=integration -coverprofile=coverage-integration.out ./internal/infrastructure/persistence/...

# E2Eテスト
test-e2e:
	@echo "Running E2E tests..."
	go test -v -race -tags=e2e -coverprofile=coverage-e2e.out ./tests/e2e/...

# 全テスト実行
test: test-unit test-integration test-e2e
	@echo "All tests completed!"

# カバレッジレポート
coverage:
	go tool cover -html=coverage-unit.out -o coverage-unit.html
	go tool cover -html=coverage-integration.out -o coverage-integration.html
	go tool cover -html=coverage-e2e.out -o coverage-e2e.html

# クリーンアップ
clean:
	rm -f coverage-*.out coverage-*.html
