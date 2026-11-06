const http = require('http');
const fs = require('fs');
const path = require('path');
const { carregarConfig, executar } = require('./gemini');

const cfg = carregarConfig();
const usuario = cfg.usuario || require('os').userInfo().username;

let sessionId = null;   // sessão do Gemini da conversa atual
let execucao = null;    // execução em andamento (só uma por vez)

function lerCorpo(req) {
  return new Promise((ok) => {
    let d = '';
    req.on('data', (c) => { d += c; if (d.length > 1e6) req.destroy(); });
    req.on('end', () => { try { ok(JSON.parse(d || '{}')); } catch { ok({}); } });
  });
}

const server = http.createServer(async (req, res) => {
  // Só aceita requisições vindas desta própria máquina/página.
  const origem = req.headers.origin;
  if (origem && !/^http:\/\/(localhost|127\.0\.0\.1)(:\d+)?$/.test(origem)) {
    res.writeHead(403); return res.end();
  }

  if (req.method === 'GET' && (req.url === '/' || req.url === '/index.html')) {
    res.writeHead(200, { 'Content-Type': 'text/html; charset=utf-8' });
    return res.end(fs.readFileSync(path.join(__dirname, 'public', 'index.html')));
  }
  
  if (req.method === 'GET' && req.url === '/api/info') {
    res.writeHead(200, { 'Content-Type': 'application/json; charset=utf-8' });
    return res.end(JSON.stringify({ usuario }));
  }

  if (req.method === 'POST' && req.url === '/api/chat') {
    const { mensagem } = await lerCorpo(req);
    if (!mensagem || !mensagem.trim()) { res.writeHead(400); return res.end(); }
    if (execucao) { res.writeHead(409); return res.end('Já existe uma execução em andamento.'); }

    res.writeHead(200, { 'Content-Type': 'application/x-ndjson; charset=utf-8', 'Cache-Control': 'no-cache' });
    const enviar = (o) => res.write(JSON.stringify(o) + '\n');

    execucao = executar(cfg, mensagem, sessionId,
      (ev) => { if (ev.type === 'init' && ev.session_id) sessionId = ev.session_id; enviar(ev); },
      (code, erro) => {
        execucao = null;
        enviar({ type: 'fim', code, erro });
        res.end();
      });
    res.on('close', () => { if (execucao) execucao.parar(); });
    return;
  }

  if (req.method === 'POST' && req.url === '/api/parar') {
    if (execucao) execucao.parar();
    res.writeHead(204); return res.end();
  }

  if (req.method === 'POST' && req.url === '/api/nova') {
    if (execucao) execucao.parar();
    sessionId = null;
    res.writeHead(204); return res.end();
  }

  res.writeHead(404); res.end();
});

server.listen(cfg.port, '127.0.0.1', () => {
  console.log(`Chat de chamados em http://localhost:${cfg.port}`);
  console.log(`Pasta de trabalho: ${cfg.workDir}`);
});
