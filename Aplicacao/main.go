// Back-end do chat de chamados: fala com a API da IA, converte o JSON do chamado
// em Markdown e o acrescenta ao final do arquivo do sistema em Historico/.
package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"chamados/dtos"
)

var (
	cfg dtos.Config
	mu  sync.Mutex // uma gravação por vez
)

func carregar() error {
	b, err := os.ReadFile("config.json")
	if err != nil {
		return errors.New("config.json não encontrado (copie config.example.json)")
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		return err
	}
	if cfg.Port == 0 {
		cfg.Port = 3000
	}
	if cfg.Provider == "" {
		cfg.Provider = "gemini"
	}
	if cfg.Model == "" {
		cfg.Model = "gemini-2.5-flash"
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}
	if cfg.APIKey == "" {
		cfg.APIKey = os.Getenv("GEMINI_API_KEY")
	}
	if cfg.Usuario == "" {
		if u, err := user.Current(); err == nil {
			cfg.Usuario = u.Username[strings.LastIndex(u.Username, `\`)+1:]
		}
	}
	// TODO: alterar esse if para, caso não exista o Context.md, perguntar se o usuário deseja criar um novo com seus parâmetros
	if _, err := os.Stat(filepath.Join(cfg.WorkDir, "Context.md")); err != nil {
		return errors.New("Context.md não encontrado em " + cfg.WorkDir)
	}
	return nil
}

func erroAPI(s int) error {
	switch s {
	case 429:
		return errors.New("Limite de uso da API atingido (Too Many Requests). Aguarde alguns minutos e tente de novo.")
	case 401, 403:
		return errors.New("A API Key foi recusada. Confira o config.json.")
	case 503:
		return errors.New("A IA está temporariamente indisponível. Tente novamente em instantes.")
	}
	return fmt.Errorf("a API da IA retornou erro %d", s)
}

// perguntar envia as instruções (Context.md) + a conversa e devolve o texto da IA.
func perguntar(ctx context.Context, instr string, hist []dtos.Msg) (string, error) {
	var url string
	var body any
	h := map[string]string{"Content-Type": "application/json"}
	if cfg.Provider == "openai" {
		ms := []map[string]string{{"role": "system", "content": instr}}
		for _, m := range hist {
			ms = append(ms, map[string]string{"role": m.Role, "content": m.Content})
		}
		url = strings.TrimRight(cfg.BaseURL, "/") + "/chat/completions"
		body = map[string]any{"model": cfg.Model, "messages": ms}
		h["Authorization"] = "Bearer " + cfg.APIKey
	} else {
		cs := []map[string]any{}
		for _, m := range hist {
			r := "user"
			if m.Role == "assistant" {
				r = "model"
			}
			cs = append(cs, map[string]any{"role": r, "parts": []map[string]string{{"text": m.Content}}})
		}
		url = "https://generativelanguage.googleapis.com/v1beta/models/" + cfg.Model + ":generateContent"
		body = map[string]any{"system_instruction": map[string]any{"parts": []map[string]string{{"text": instr}}}, "contents": cs}
		h["x-goog-api-key"] = cfg.APIKey
	}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(b))
	for k, v := range h {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		return "", errors.New("não foi possível contatar a API da IA")
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		log.Printf("API da IA respondeu %d", resp.StatusCode) // nunca registra a chave
		return "", erroAPI(resp.StatusCode)
	}
	var out struct {
		Candidates []struct{ Content struct{ Parts []struct{ Text string } } }
		Choices    []struct{ Message struct{ Content string } }
	}
	json.Unmarshal(raw, &out)
	var t strings.Builder
	if len(out.Candidates) > 0 {
		for _, p := range out.Candidates[0].Content.Parts {
			t.WriteString(p.Text)
		}
	} else if len(out.Choices) > 0 {
		t.WriteString(out.Choices[0].Message.Content)
	}
	if t.Len() == 0 {
		return "", errors.New("a IA retornou uma resposta vazia")
	}
	return t.String(), nil
}

// ordenado lê o objeto JSON do chamado mantendo a ordem dos campos.
func ordenado(raw []byte) ([]dtos.Campo, error) {
	d := json.NewDecoder(bytes.NewReader(raw))
	if t, err := d.Token(); err != nil || t != json.Delim('{') {
		return nil, errors.New("o chamado não é um objeto JSON")
	}
	var cs []dtos.Campo
	for d.More() {
		k, _ := d.Token()
		var v any
		if err := d.Decode(&v); err != nil {
			return nil, err
		}
		cs = append(cs, dtos.Campo{k.(string), v})
	}
	return cs, nil
}

// extrairTicket procura na resposta da IA um objeto JSON com o campo "sistema".
func extrairTicket(t string) json.RawMessage {
	i, j := strings.Index(t, "{"), strings.LastIndex(t, "}")
	if i < 0 || j < i {
		return nil
	}
	cs, err := ordenado([]byte(t[i : j+1]))
	if err != nil {
		return nil
	}
	for _, c := range cs {
		if c.K == "sistema" {
			return json.RawMessage(t[i : j+1])
		}
	}
	return nil
}

// ticketParaMD converte o chamado em Markdown
func ticketParaMD(id string, cs []dtos.Campo) string {
	var b strings.Builder
	fmt.Fprintf(&b, "\n\n---\n\n## Chamado %s — %s — %s\n\n", id[:8], time.Now().Format("02/01/2006 15:04"), cfg.Usuario)
	for _, c := range cs {
		if c.K == "sistema" {
			continue
		}
		switch v := c.V.(type) {
		case string:
			fmt.Fprintf(&b, "**%s:** %s\n\n", c.K, v)
		case []any:
			fmt.Fprintf(&b, "**%s:**\n", c.K)
			for _, x := range v {
				fmt.Fprintf(&b, "- %v\n", x)
			}
			b.WriteString("\n")
		default:
			j, _ := json.Marshal(v)
			fmt.Fprintf(&b, "**%s:** %s\n\n", c.K, j)
		}
	}
	return b.String()
}

func salvar(id string, raw json.RawMessage) (string, error) {
	cs, err := ordenado(raw)
	if err != nil {
		return "", err
	}
	sistema := ""
	for _, c := range cs {
		if c.K == "sistema" {
			sistema, _ = c.V.(string)
		}
	}
	dir := filepath.Join(cfg.WorkDir, "Historico")
	arq := ""
	ents, _ := os.ReadDir(dir)
	for _, e := range ents { // só grava em arquivos que já existem (evita nomes inventados)
		if n := e.Name(); strings.HasSuffix(n, ".md") && strings.EqualFold(strings.TrimSuffix(n, ".md"), sistema) {
			arq = filepath.Join(dir, n)
		}
	}
	if arq == "" {
		return "", fmt.Errorf("sistema %q não tem arquivo em Historico", sistema)
	}
	nome := filepath.Base(arq)

	mu.Lock()
	defer mu.Unlock()
	reg := filepath.Join(cfg.WorkDir, "Dados", "tickets.jsonl")
	if b, _ := os.ReadFile(reg); bytes.Contains(b, []byte(`"id":"`+id+`"`)) {
		return nome, nil // já gravado: não duplica
	}
	bk := filepath.Join(cfg.WorkDir, "Backups", strings.TrimSuffix(nome, ".md")+"-"+time.Now().Format("20060102")+".md")
	if _, err := os.Stat(bk); err != nil { // um backup por dia, antes da primeira gravação
		if b, e := os.ReadFile(arq); e == nil {
			os.MkdirAll(filepath.Dir(bk), 0o755)
			os.WriteFile(bk, b, 0o644)
		}
	}
	f, err := os.OpenFile(arq, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return "", err
	}
	_, err = f.WriteString(ticketParaMD(id, cs))
	f.Close()
	if err != nil {
		return "", err
	}
	var cb bytes.Buffer
	json.Compact(&cb, raw)
	linha, _ := json.Marshal(map[string]any{"id": id, "usuario": cfg.Usuario, "criado_em": time.Now().Format(time.RFC3339), "arquivo": nome, "ticket": json.RawMessage(cb.Bytes())})
	os.MkdirAll(filepath.Dir(reg), 0o755)
	j, err := os.OpenFile(reg, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nome, err
	}
	defer j.Close()
	j.Write(append(linha, '\n'))
	log.Printf("chamado %s gravado em %s", id, nome)
	return nome, nil
}

func novoID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6], b[8] = b[6]&0x0f|0x40, b[8]&0x3f|0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func jsonOut(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(v)
}

func local(h http.Handler) http.Handler { // só aceita páginas desta máquina
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if o := r.Header.Get("Origin"); o != "" && !strings.HasPrefix(o, "http://localhost") && !strings.HasPrefix(o, "http://127.0.0.1") {
			http.Error(w, "origem negada", 403)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func main() {
	if err := carregar(); err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/info", func(w http.ResponseWriter, r *http.Request) {
		jsonOut(w, map[string]string{"usuario": cfg.Usuario})
	})
	mux.HandleFunc("POST /api/chat", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Mensagens []dtos.Msg `json:"mensagens"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in) != nil || len(in.Mensagens) == 0 {
			http.Error(w, "requisição inválida", 400)
			return
		}
		instr, err := os.ReadFile(filepath.Join(cfg.WorkDir, "Context.md")) // relido a cada pedido
		if err != nil {
			http.Error(w, "Context.md não encontrado", 500)
			return
		}
		txt, err := perguntar(r.Context(), string(instr), in.Mensagens)
		if err != nil {
			http.Error(w, err.Error(), 502)
			return
		}
		out := map[string]any{"mensagem": txt}
		if t := extrairTicket(txt); t != nil {
			out["ticket"], out["id"] = t, novoID()
		}
		jsonOut(w, out)
	})
	mux.HandleFunc("POST /api/tickets", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			ID     string          `json:"id"`
			Ticket json.RawMessage `json:"ticket"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in) != nil || len(in.ID) != 36 {
			http.Error(w, "requisição inválida", 400)
			return
		}
		arq, err := salvar(in.ID, in.Ticket)
		if err != nil {
			http.Error(w, err.Error(), 422)
			return
		}
		jsonOut(w, map[string]string{"arquivo": arq})
	})
	mux.Handle("/", http.FileServer(http.Dir("web/dist"))) // front compilado

	addr := fmt.Sprintf("127.0.0.1:%d", cfg.Port)
	log.Printf("Chat de chamados em http://localhost:%d (pasta: %s)", cfg.Port, cfg.WorkDir)
	log.Fatal(http.ListenAndServe(addr, local(mux)))
}
