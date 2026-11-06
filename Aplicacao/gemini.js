// Executa o Gemini CLI como processo filho e entrega os eventos (stream-json).
const { spawn, execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const WIN = process.platform === 'win32';

function carregarConfig() {
  const arq = path.join(__dirname, 'config.json');
  if (!fs.existsSync(arq)) {
    throw new Error('config.json não encontrado. Copie config.example.json para config.json e ajuste.');
  }
  const cfg = JSON.parse(fs.readFileSync(arq, 'utf8'));
  cfg.port = cfg.port || 3000;
  if (!cfg.workDir || !fs.existsSync(cfg.workDir)) {
    throw new Error('workDir inválido em config.json: ' + cfg.workDir);
  }
  if (!fs.existsSync(path.join(cfg.workDir, 'GEMINI.md'))) {
    console.warn('Aviso: GEMINI.md não encontrado em ' + cfg.workDir);
  }
  return cfg;
}

// A mensagem do usuário vai pela entrada padrão (stdin), nunca na linha de comando.
// Assim não há problema de aspas nem risco de injeção de comandos no Windows.
function executar(cfg, mensagem, sessionId, onEvento, onFim) {
  const args = ['-p', 'Responda', '--output-format', 'stream-json', '--approval-mode', 'plan'];
  if (sessionId) args.push('--resume', sessionId);
  if (cfg.model) args.push('-m', cfg.model);
  if (cfg.skipTrust) args.push('--skip-trust');

  const env = { ...process.env };
  if (cfg.apiKey) env.GEMINI_API_KEY = cfg.apiKey; // só no processo filho

  const proc = spawn('gemini', args, { cwd: cfg.workDir, env, shell: WIN, windowsHide: true });
  let buffer = '';
  let erro = '';
  let encerrado = false;

  const limpar = (t) => (cfg.apiKey ? t.split(cfg.apiKey).join('***') : t);

  proc.stdout.setEncoding('utf8');
  proc.stdout.on('data', (chunk) => {
    buffer += chunk;
    let i;
    while ((i = buffer.indexOf('\n')) >= 0) {
      const linha = buffer.slice(0, i).trim();
      buffer = buffer.slice(i + 1);
      if (!linha) continue;
      try { onEvento(JSON.parse(linha)); }
      catch { onEvento({ type: 'texto_bruto', content: limpar(linha) }); }
    }
  });
  proc.stderr.setEncoding('utf8');
  proc.stderr.on('data', (c) => { erro = (erro + c).slice(-2000); });
  proc.on('error', (e) => {
    if (encerrado) return; encerrado = true;
    onFim(-1, 'Não foi possível iniciar o gemini: ' + e.message);
  });
  proc.on('close', (code) => {
    if (encerrado) return; encerrado = true;
    onFim(code, code ? limpar(erro) : '');
  });

  proc.stdin.on('error', () => {});
  proc.stdin.end(mensagem);

  return {
    parar() {
      try {
        if (WIN) execSync(`taskkill /pid ${proc.pid} /T /F`, { stdio: 'ignore' });
        else proc.kill('SIGTERM');
      } catch {}
    },
  };
}

module.exports = { carregarConfig, executar };
