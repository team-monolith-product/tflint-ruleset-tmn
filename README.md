# tflint-ruleset-tmn

[![Build Status](https://github.com/team-monolith-product/tflint-ruleset-tmn/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/team-monolith-product/tflint-ruleset-tmn/actions)

팀모노리스 Terraform 리포지토리를 위한 커스텀 TFLint 룰셋입니다.

## Requirements

- TFLint v0.46+
- Go v1.22+ (빌드 시)

## Installation

`.tflint.hcl` 파일에 다음을 추가합니다:

```hcl
plugin "tmn" {
  enabled = true

  version = "0.1.0"
  source  = "github.com/team-monolith-product/tflint-ruleset-tmn"
}
```

설치:

```bash
tflint --init
```

## Rules

| Name | Description | Severity | Enabled |
| --- | --- | --- | --- |
| aws_instance_example_type | EC2 인스턴스 타입 검증 | ERROR | ✔ |
| aws_s3_bucket_example_lifecycle_rule | S3 버킷 lifecycle 룰 검증 | ERROR | ✔ |
| google_compute_ssl_policy | GCP SSL 정책 검증 | WARNING | ✔ |
| terraform_backend_type | Terraform backend 타입 검증 | ERROR | ✔ |

## Building

### 로컬 빌드

```bash
# 빌드
make build

# 테스트
make test

# 로컬 설치 (~/.tflint.d/plugins)
make install
```

### 로컬 테스트

```bash
cat << EOS > .tflint.hcl
plugin "tmn" {
  enabled = true
}
EOS

tflint
```

## Release

| Branch | 동작 |
| --- | --- |
| `develop` | Push 시 `v0.1.0-dev.<sha>` pre-release 자동 생성 |
| `main` | Push 시 `main.go`의 버전으로 정식 릴리즈 자동 생성 |

### 새 버전 릴리즈

1. `main.go`의 `Version` 값을 변경
2. `main` 브랜치에 병합
3. 자동으로 태그 생성 및 릴리즈

## 새 룰 추가하기

1. `rules/` 디렉토리에 새 룰 파일 생성
2. `main.go`의 `Rules` 슬라이스에 룰 등록
3. 테스트 작성 (`_test.go`)

참고: [TFLint Custom Rules Guide](https://github.com/terraform-linters/tflint/blob/master/docs/developer-guide/plugins.md)
