# System Design - Modulos Externos no Cultivator

## Visao Geral

O Cultivator foi estendido para suportar **modulos Terraform externos** hospedados em repositorios Git (GitHub, GitLab) ou disponiveis via HTTP/HTTPS. Este documento descreve a arquitetura, padroes de design e fluxos de implementação.

## Objetivos

1. **Deteccao automatica** de mudancas em modulos externos
2. **Download e checkout** inteligente de modulos antes de plan/apply
3. **Suporte multiplos formatos**: `git::`, `https://`, `http://`
4. **Arquitetura extensivel** para adicionar novos tipos de fontes
5. **Respeito aos principios SOLID e Clean Code**

## Arquitetura Geral

```
┌─────────────────────────────────────────────────────────┐
│         GitHub Event (PR aberto/atualizado)             │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│              Orchestrator (main logic)                  │
└────────────────────┬────────────────────────────────────┘
                     │
         ┌───────────┼───────────┐
         ▼           ▼           ▼
┌──────────────┐ ┌────────┐ ┌──────────┐
│   Detector   │ │ Parser │ │ Executor │
│              │ │        │ │          │
│ Detects:     │ │Parses: │ │Executes: │
│ • Local      │ │ • HCL  │ │ • Plans  │
│   changes    │ │ • Ext. │ │ • Applies│
│ • External   │ │  sources│ │• Checkout│
│   sources    │ │        │ │          │
└──────┬───────┘ └────┬───┘ └────┬─────┘
       │              │          │
       └──────┬───────┴──────┬───┘
              │              │
              ▼              ▼
      ┌───────────────────────────────┐
      │     SourceParser (Strategy)   │
      ├───────────────────────────────┤
      │  • Detect source type         │
      │  • Delegate to correct impl.  │
      └───────────┬───────────────────┘
                  │
      ┌───────────┴───────────┐
      ▼                       ▼
┌──────────────┐      ┌──────────────┐
│ GitModule    │      │ HTTPModule   │
│ Source       │      │ Source       │
├──────────────┤      ├──────────────┤
│ Parse()      │      │ Parse()      │
│ FetchVersion │      │ FetchVersion │
│ Checkout()   │      │ Checkout()   │
│              │      │              │
│ • git clone  │      │ • Download   │
│ • git fetch  │      │ • Extract    │
│ • git co     │      │   (tar/zip)  │
└──────────────┘      └──────────────┘
```

## Componentes Principais

### 1. ModuleSource Interface (Abstração)

**Arquivo**: `pkg/module/source.go`

Define o contrato para qualquer fonte de módulo:

```go
type ModuleSource interface {
    Parse(source string) (*SourceInfo, error)
    FetchVersion(ctx context.Context, source string) (string, error)
    Checkout(ctx context.Context, source string, workdir string) error
    Type() string
}
```

**Beneficios SOLID**:
- **Interface Segregation**: Interface enxuta com apenas metodos essenciais
- **Dependency Inversion**: Depende de abst racoes, nao de implementacoes

### 2. GitModuleSource (Strategy Pattern)

**Arquivo**: `pkg/module/git.go`

Implementa suporte para módulos Git:

```
Formatos suportados:
├── git::https://github.com/org/repo//vpc?ref=v1.0.0
├── git::https://github.com/org/repo//database?ref=main
├── git::ssh://git@gitlab.com/org/repo.git//app?ref=HEAD
└── git::https://github.com/org/repo//nested/path
```

**Operações**:
- `Parse()`: Extrai URL, subpath, ref
- `FetchVersion()`: Usa `git ls-remote` (sem clonar)
- `Checkout()`: `git clone` + `git checkout`

**DRY Principle**:
```go
// Métodos extracted por responsabilidade:
gitClone()              // Clone operation
extractRepositoryURL()  // URL parsing
```

### 3. HTTPModuleSource (Strategy Pattern)

**Arquivo**: `pkg/module/http.go`

Implementa suporte para módulos HTTP/HTTPS:

```
Formatos suportados:
├── https://github.com/org/repo/archive/refs/tags/v1.0.0.tar.gz
├── https://github.com/org/repo/releases/download/v1.0.0/module.zip
├── https://registry.terraform.io/modules/path/module.tar.gz
└── https://example.com/modules/vpc/v2.0.0.tar.gz
```

**Operações**:
- `Parse()`: Extrai URL, extrai ref do arquivo
- `FetchVersion()`: HEAD request + ETag ou hash da URL
- `Checkout()`: Download + Extract (tar.gz ou zip)

**DRY Principle**:
```go
// Métodos extracted:
downloadFile()          // Download operation
extractTarGz()          // TAR extraction
extractZip()            // ZIP extraction
shouldExtractFile()     // Filter logic
getTargetPath()         // Path calculation
writeFile()             // File writing
```

### 4. SourceParser (Facade + Factory)

**Arquivo**: `pkg/module/source.go`

Orquestra a seleção da implementação correta:

```go
parser := NewSourceParser()

// Auto-detect tipo e delegar
info, err := parser.Parse("git::https://github.com/org/repo//vpc?ref=v1.0.0")
// → Detecta "git::" → Usa GitModuleSource

info, err := parser.Parse("https://github.com/org/repo/archive/refs/tags/v1.0.0.tar.gz")
// → Detecta "https://" → Usa HTTPModuleSource
```

**Benefícios SOLID**:
- **Open/Closed**: Facil adicionar novo tipo sem modificar SourceParser
- **Single Responsibility**: Apenas seleciona implementacao correta

### 5. Parser Estendido

**Arquivo**: `pkg/parser/parser.go`

Estendido para extrair módulos externos:

```go
// Campo novo em TerraformBlock
type TerraformBlock struct {
    Source       string              // "local" ou "git::" ou "https://"
    ExternalInfo *module.SourceInfo  // Apenas se for external
    IsExternal   bool                // Flag para facilitar checks
}

// Método novo
func (p *Parser) GetExternalModuleSources(modulePath string) ([]module.SourceInfo, error)
```

### 6. Detector Estendido

**Arquivo**: `pkg/detector/detector.go`

Detecta **dois tipos de mudanças**:

#### 6a. Mudanças Locais
```go
DetectChangedModules() // Já existia
// Detecta: arquivo .tf mudou, terragrunt.hcl mudou, etc
```

#### 6b. Mudanças em Módulos Externos
```go
DetectExternalModuleChanges(ctx context.Context, modules []Module) ([]Module, error)
// Novo método!
// Detecta: ref mudou de v1.0.0 → v1.1.0
```

**Fluxo**:
1. Obtém terragrunt.hcl do commit base
2. Obtém terragrunt.hcl do commit head
3. Compara sources de módulos externos
4. Se ref/URL mudou → módulo mudou

```go
// Exemplo de detecção:
base:   terraform { source = "git::https://github.com/org/repo//vpc?ref=v1.0.0" }
head:   terraform { source = "git::https://github.com/org/repo//vpc?ref=v1.1.0" }
result: Mudanca detectada (ref: v1.0.0 → v1.1.0)
```

### 7. Executor Estendido

**Arquivo**: `pkg/executor/executor.go`

Adiciona capacidade de checkout:

```go
// Campo novo
type Executor struct {
    // ... campos existentes
    sourceParser *module.SourceParser
}

// Novo método (a implementar em próximas versões)
func (e *Executor) PrepareExternalModules(ctx context.Context, sources []module.SourceInfo) error
```

**Fluxo de execução**:
1. Executor.Plan(module) é chamado
2. Antes de rodar terragrunt plan:
   - Detecta módulos externos no terragrunt.hcl
   - Faz checkout via SourceParser
   - Aí sim executa plan

## Fluxo de Dados (End-to-End)

### Cenário: PR atualiza referência de módulo externo

```
1. PR aberto/atualizado
   └─ Novo commit em: environments/prod/terragrunt.hcl
   
2. GitHub event → Cultivator
   └─ base: abc123 (commit anterior)
   └─ head: def456 (commit novo)

3. Detector.DetectExternalModuleChanges()
   ├─ git show abc123:environments/prod/terragrunt.hcl
   │  └─ source = "git::https://github.com/org/modules//vpc?ref=v1.0.0"
   │
   ├─ git show def456:environments/prod/terragrunt.hcl
   │  └─ source = "git::https://github.com/org/modules//vpc?ref=v1.1.0"
   │
   └─ Compara: v1.0.0 != v1.1.0 -> MUDANCA DETECTADA

4. SourceParser.Parse("git::https://github.com/org/modules//vpc?ref=v1.1.0")
   └─ Detecta prefixo "git::" → Usa GitModuleSource

5. GitModuleSource.Parse()
   └─ Extrai: {
        URL: "https://github.com/org/modules",
        Ref: "v1.1.0",
        SubPath: "/vpc"
      }

6. Executor.Plan()
   ├─ GitModuleSource.Checkout(...)
   │  ├─ git clone --depth 1 --branch v1.1.0 https://github.com/org/modules /tmp/cultivator-module
   │  └─ Modulo pronto
   │
   └─ terragrunt plan
      └─ Usa o módulo já clonado em /tmp

7. Formata resultado e comenta na PR
   └─ "Modules changed: environments/prod (vpc source: v1.0.0 -> v1.1.0)"
```

## Padroes de Design

### 1. Strategy Pattern
Implementações diferentes para Git vs HTTP:
```
ModuleSource (Interface)
  ├── GitModuleSource
  └── HTTPModuleSource
```

**Benefício**: Fácil adicionar novo tipo (S3, Artifactory, etc)

### 2. Factory Pattern
SourceParser cria a estratégia correta:
```go
parser.Parse("git::...")      // → GitModuleSource
parser.Parse("https://...")   // → HTTPModuleSource
parser.Parse("s3://...")      // → Nova implementação futura
```

### 3. Facade Pattern
SourceParser esconde complexidade de múltiplas implementações

### 4. Dependency Injection
```go
// Injetar SourceParser no Executor
executor := NewExecutor(workdir, stdout, stderr)
// sourceParser já criado e injetado no constructor
```

## Seguranca

### 1. Validação de URLs
```go
// Apenas URLs válidas
func isValidURL(urlStr string) error {
    _, err := url.Parse(urlStr)
    return err
}
```

### 2. Timeout em Operações
```go
// Todas operações recebem context.Context com timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

err := moduleSource.Checkout(ctx, source, workdir)
```

### 3. Sanitização de Paths
```go
// Evita path traversal
targetPath := filepath.Join(workdir, cleanedPath)
```

## Escalabilidade

### Adicionar novo tipo de fonte

**Passo 1**: Implementar interface
```go
type S3ModuleSource struct {
    // ...
}

func (s *S3ModuleSource) Parse(source string) (*SourceInfo, error) { }
func (s *S3ModuleSource) FetchVersion(ctx context.Context, source string) (string, error) { }
func (s *S3ModuleSource) Checkout(ctx context.Context, source string, workdir string) error { }
func (s *S3ModuleSource) Type() string { return "s3" }
```

**Passo 2**: Registrar em SourceParser
```go
func NewSourceParser() *SourceParser {
    return &SourceParser{
        sources: map[string]ModuleSource{
            "git":  NewGitModuleSource(),
            "http": NewHTTPModuleSource(),
            "s3":   NewS3ModuleSource(),     // NOVO
        },
    }
}
```

**Passo 3**: Atualizar detection logic
```go
func (sp *SourceParser) detectSourceType(source string) string {
    if strings.HasPrefix(source, "s3://") {
        return "s3"
    }
    // ... resto do código
}
```

Nenhuma outra mudanca necessaria! Respeita Open/Closed Principle.

## Testabilidade

### Unit Tests
```go
// Cada implementação pode ser testada isoladamente
TestGitModuleSource_Parse()
TestHTTPModuleSource_Parse()
TestSourceParser_DetectSourceType()

// Mock de ModuleSource para testar Executor
type MockModuleSource struct { }
```

### Integration Tests (futuro)
```go
// Testar fluxo completo
1. Criar repo git temporário
2. Criar archive HTTP temporário
3. Executar Detector + SourceParser + Executor
4. Verificar checkout correto
```

## Roadmap de Implementação

### Phase 1 (Complete)
- [x] Interface ModuleSource
- [x] GitModuleSource
- [x] HTTPModuleSource
- [x] SourceParser
- [x] Parser extension
- [x] Detector extension
- [x] Unit tests

### 🚧 Phase 2 (Próximo)
- [ ] Executor.PrepareExternalModules()
- [ ] Integração com Orchestrator
- [ ] Cache de módulos externos
- [ ] Integration tests

### Phase 3 (Future)
- [ ] Suporte S3
- [ ] Suporte Artifactory
- [ ] Suporte OCI Registry
- [ ] Web UI para visualizar módulos
- [ ] Drift detection para módulos externos

## Usage Examples

### Exemplo 1: Git com subpath e ref

**terragrunt.hcl**:
```hcl
terraform {
  source = "git::https://github.com/acme-corp/terraform-modules//networking/vpc?ref=v2.3.1"
}

inputs = {
  vpc_cidr = "10.0.0.0/16"
  environment = "production"
}
```

**O que acontece**:
1. Detecta modulo externo
2. Faz checkout: `git clone --branch v2.3.1 https://github.com/acme-corp/terraform-modules`
3. Usa arquivo: `networking/vpc/main.tf` (subpath)
4. Roda plan com inputs customizados

### Exemplo 2: HTTP archive

**terragrunt.hcl**:
```hcl
terraform {
  source = "https://github.com/terraform-aws-modules/terraform-aws-s3-bucket/releases/download/v3.8.0/terraform-aws-s3-bucket-v3.8.0.tar.gz#s3-bucket"
}
```

**O que acontece**:
1. Detecta HTTP source
2. Download: `wget https://github.com/terraform-aws-modules/...`
3. Extrai: `tar -xzf terraform-aws-s3-bucket-v3.8.0.tar.gz`
4. Usa subpath: `s3-bucket/` (apos #)
5. Roda plan

### Exemplo 3: Mudança de versão detectada automaticamente

**Commit anterior**:
```hcl
source = "git::https://github.com/acme/modules//vpc?ref=v1.0.0"
```

**Novo commit (na PR)**:
```hcl
source = "git::https://github.com/acme/modules//vpc?ref=v1.1.0"
```

**Resultado no PR**:
```
Auto-plan detected changes

Modules changed: environments/production
  └─ vpc: external source updated (v1.0.0 → v1.1.0)

Plan: 2 additions, 1 modification, 0 destructions
```

## Debugging

### Logs úteis para troubleshoot

```bash
# Ver qual SourceParser foi selecionado
DEBUG=true cultivator run

# Output esperado:
# "Detecting source type for: git::https://..."
# "Selected handler: GitModuleSource"
# "Parsed info: {URL: ..., Ref: v1.1.0, SubPath: /vpc}"

# Ver operações de checkout
# "Downloading module from https://..."
# "Extracting archive to /tmp/cultivator-module-abc123"
# "Module ready at /tmp/cultivator-module-abc123/vpc"
```

## Referencias

### Padrões SOLID Aplicados

| Princípio | Aplicação |
|-----------|-----------|
| **S - Single Responsibility** | Cada ModuleSource implementação tem 1 responsabilidade |
| **O - Open/Closed** | Novo tipo = nova implementação, sem alterar existing |
| **L - Liskov Substitution** | Qualquer ModuleSource funciona igual (contrato) |
| **I - Interface Segregation** | Interface enxuta, sem métodos desnecessários |
| **D - Dependency Inversion** | Depende de ModuleSource (abstração), não de Git/HTTP |

### Princípios Clean Code

- Nomes descritivos (`extractRefFromURL`, `shouldExtractFile`)
- Funcoes pequenas (max ~30 linhas)
- Metodos extracted para DRY (`gitClone`, `downloadFile`, `writeFile`)
- Comentarios explicam WHY, nao WHAT
- Error handling explicito

### Referências Externas

- [Strategy Pattern](https://refactoring.guru/design-patterns/strategy)
- [Factory Pattern](https://refactoring.guru/design-patterns/factory-method)
- [SOLID Principles](https://en.wikipedia.org/wiki/SOLID)
- [Clean Code by Robert Martin](https://www.oreilly.com/library/view/clean-code-a/9780136083238/)
