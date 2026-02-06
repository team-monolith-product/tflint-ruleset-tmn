# tflint-ruleset-tmn

[![Build Status](https://github.com/team-monolith-product/tflint-ruleset-tmn/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/team-monolith-product/tflint-ruleset-tmn/actions)

팀모노리스 Terraform 리포지토리를 위한 커스텀 TFLint 룰셋입니다.

**적용 대상:** [ped-terraform](https://github.com/team-monolith-product/ped-terraform)

## Requirements

- TFLint v0.47+ (autofix 지원)
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
| foreach_toset | `{ for x in list : x => x }` 패턴을 `toset()`으로 변환 (autofix 지원) | WARNING | ✔ |
| sort_provider_app_map | `provider_to_app_map` 내 키 알파벳순 정렬 및 불필요한 공백 제거 (autofix 지원) | WARNING | ✔ |
| sort_additional_secrets_data | `additional_secrets[*].data` 내 키 알파벳순 정렬 (autofix 지원) | WARNING | ✔ |
| require_additional_parameters_trailing_comma | `additional_parameters` 배열 마지막 요소 뒤에 trailing comma 요구 (autofix 지원) | WARNING | ✔ |

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
| `develop` | Push 시 `v0.0.0-dev` pre-release 업데이트 |
| `main` | Push 시 `main.go`의 버전으로 정식 릴리즈 자동 생성 |

### 새 버전 릴리즈

1. `main.go`의 `Version` 값을 변경
2. `main` 브랜치에 병합
3. 자동으로 태그 생성 및 릴리즈

## 새 룰 추가하기

1. `rules/` 디렉토리에 룰 파일 생성
2. `tests/` 디렉토리에 테스트 파일 생성
3. `main.go`의 `Rules` 슬라이스에 룰 등록
4. `README.md`의 Rules 테이블 업데이트
5. dev 버전으로 ped-terraform 검증 (아래 절차 참고)
6. `main.go`의 `Version` 값을 올림

**규칙:**
- 룰 이름은 동사형으로 작성 (예: `sort_xxx`, `validate_xxx`, `require_xxx`)
- 룰 파일 상단에 `@example GOOD`과 `@example BAD` 주석을 반드시 포함

**룰 파일 주석 형식:**
```go
// Rule: rule_name
//
// 룰에 대한 설명
//
// @example GOOD
// // 올바른 코드 예시
//
// @example BAD
// // 잘못된 코드 예시

package rules
```

### 새 룰 검증 (ped-terraform)

새 룰 추가 시 반드시 ped-terraform에서 dev 버전으로 검증한 후 main에 병합합니다.

1. tflint-ruleset-tmn의 feature 브랜치를 develop에 병합하여 `v0.0.0-dev` 릴리즈를 트리거합니다.
2. ped-terraform에서 검증용 브랜치를 생성합니다.
3. `.tflint.hcl`을 dev 버전으로 변경합니다:
   ```hcl
   plugin "tmn" {
     enabled = true
     version = "0.0.0-dev"
     source  = "github.com/team-monolith-product/tflint-ruleset-tmn"
   }
   ```
4. `tflint --init`으로 dev 버전을 설치합니다.
5. `tflint`을 실행하여 린트 결과를 확인합니다.
6. autofix 지원 룰의 경우 `tflint --fix`를 실행하여 수정 결과를 확인합니다.
7. ped-terraform에 draft PR을 생성하여 린트 결과를 리뷰합니다.
8. 검증이 완료되면 tflint-ruleset-tmn에서 main으로 PR을 생성합니다.
9. tflint-ruleset-tmn 정식 릴리즈 후, ped-terraform draft PR의 `.tflint.hcl` 버전을 정식 버전으로 변경하여 린트 버전 업그레이드 PR로 사용합니다.

참고: [TFLint Custom Rules Guide](https://github.com/terraform-linters/tflint/blob/master/docs/developer-guide/plugins.md)
