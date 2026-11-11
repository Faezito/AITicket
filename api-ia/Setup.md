# Swagger — Documentação da API

Esta API utiliza **Swaggo + Swagger UI** para gerar e disponibilizar uma documentação interativa dos endpoints.

A documentação permite visualizar os endpoints, consultar parâmetros, verificar respostas e executar requisições diretamente pelo navegador.

## 1. Tecnologias utilizadas

- [Swaggo](https://github.com/swaggo/swag): geração automática da documentação.
- [Swagger UI](https://swagger.io/tools/swagger-ui/): interface web interativa.
- [HTTP Swagger](https://github.com/swaggo/http-swagger): integração do Swagger UI com o servidor Go.
- OpenAPI: especificação da documentação da API.

## 2. Instalação em outro ambiente

Certifique-se de que o [Go](https://go.dev/dl/) esteja instalado.

### 2.1. Instalar o Swaggo CLI

Execute no PowerShell:

```powershell
go install github.com/swaggo/swag/cmd/swag@v1.16.4
```

Verifique a instalação:

```powershell
swag --version
```

### 2.2. Instalar as dependências

Na raiz do projeto (`api-ia`), execute:

```powershell
go get github.com/swaggo/http-swagger
go get github.com/swaggo/swag@v1.16.4
go mod tidy
```

> **Importante:** mantenha a versão do Swaggo CLI e da dependência `github.com/swaggo/swag` alinhadas para evitar incompatibilidades na geração da documentação.

## 3. Configuração do Swagger

### 3.1. Metadados da API

No arquivo `main.go`, adicione os comentários de configuração antes dos imports:

```go
// @title           AI Ticket API
// @version         1.0
// @description     API para comunicação com modelos de IA locais através do Ollama.
// @termsOfService  http://swagger.io/terms/

// @contact.name   Matheus
// @license.name   MIT

// @host      localhost:8080
// @BasePath  /
```

### 3.2. Imports

Adicione os seguintes imports:

```go
_ "api-ia/docs"
httpSwagger "github.com/swaggo/http-swagger"
```

O import `api-ia/docs` registra a especificação gerada pelo Swaggo.

### 3.3. Registrar a rota

Dentro da função `main()`, após criar o `http.ServeMux`, registre:

```go
mux.Handle("GET /swagger/", httpSwagger.Handler())
```

## 4. Documentação dos endpoints

Para documentar um endpoint, adicione comentários especiais imediatamente acima da implementação do handler.

Exemplo:

```go
// handleAsk godoc
// @Summary      Enviar pergunta para IA
// @Description  Envia uma pergunta para o modelo local através do Ollama.
// @Tags         AI
// @Accept       json
// @Produce      json
// @Param        request body dtos.AskRequest true "Dados da solicitação"
// @Success      200 {object} dtos.AskResponse "Resposta gerada com sucesso"
// @Failure      400 {object} dtos.ErrorResponse "JSON inválido ou campos obrigatórios ausentes"
// @Failure      409 {object} dtos.ErrorResponse "Request ID já está em uso"
// @Failure      499 {object} dtos.ErrorResponse "Geração cancelada"
// @Failure      502 {object} dtos.ErrorResponse "Erro na comunicação com Ollama"
// @Failure      504 {object} dtos.ErrorResponse "Tempo limite excedido"
// @Router       /api/ask [post]
func (s *Server) handleAsk(w http.ResponseWriter, r *http.Request) {
    // Implementação do endpoint
}
```

### Principais anotações

| Anotação | Descrição |
|---|---|
| `@Summary` | Resumo do endpoint |
| `@Description` | Descrição detalhada |
| `@Tags` | Categoria do endpoint |
| `@Accept` | Formato de entrada |
| `@Produce` | Formato de saída |
| `@Param` | Parâmetros da requisição |
| `@Success` | Resposta de sucesso |
| `@Failure` | Respostas de erro |
| `@Router` | Caminho e método HTTP |

## 5. Gerar ou atualizar a documentação

Sempre que adicionar ou modificar as anotações de um endpoint, execute na raiz do projeto:

```powershell
swag init
```

O comando gera ou atualiza automaticamente a pasta `docs/`:

```text
api-ia/
├── docs/
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── main.go
├── go.mod
├── go.sum
└── ...
```

**Importante:** mantenha a pasta `docs/` versionada no Git. Ela contém os arquivos necessários para compilar a API com a documentação integrada, mesmo em ambientes que não possuem o Swaggo CLI instalado.

## 6. Executar a API

Inicie a aplicação:

```powershell
go run .
```

Acesse a interface Swagger UI pelo navegador:

**http://localhost:8080/swagger/index.html**

A partir da interface, é possível consultar os endpoints e executar requisições utilizando o recurso `Try it out`.

## 7. Fluxo de atualização

Sempre que houver alterações na documentação dos endpoints:

1. Atualize as anotações no código Go.
2. Execute `swag init`.
3. Verifique os arquivos gerados na pasta `docs/`.
4. Execute a API.
5. Confira a documentação pelo navegador.

---

**Projeto:** AI Ticket API  
**Tecnologia:** Go  
**Integração de IA:** Ollama local  
**Documentação:** Swagger UI / OpenAPI