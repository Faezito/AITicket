const $ = (id) => document.getElementById(id);
const col = $('col'), msgs = $('msgs'), txt = $('txt');
let ocupado = false;

function add(cls, texto) {
  $('vazio')?.remove();
  const d = document.createElement('div');
  d.className = 'm ' + cls; d.textContent = texto || '';
  col.appendChild(d); msgs.scrollTop = msgs.scrollHeight; return d;
}
function estado(on, msg) {
  ocupado = on; $('enviar').style.display = on ? 'none' : '';
  $('parar').style.display = on ? '' : 'none'; $('status').textContent = msg || '';
}
txt.addEventListener('input', () => { txt.style.height = 'auto'; txt.style.height = txt.scrollHeight + 'px'; });
txt.addEventListener('keydown', (e) => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); enviar(); } });
$('enviar').onclick = enviar;
$('parar').onclick = () => fetch('/api/parar', { method: 'POST' });
$('nova').onclick = async () => {
  await fetch('/api/nova', { method: 'POST' });
  col.innerHTML = '<p class="vazio" id="vazio">Descreva o atendimento para começar.</p>'; estado(false);
};

async function enviar() {
  const mensagem = txt.value.trim();
  if (!mensagem || ocupado) return;
  txt.value = ''; txt.style.height = 'auto';
  add('u', mensagem); estado(true, 'Gemini está processando…');
  let resp = null, interrompido = true;
  try {
    const r = await fetch('/api/chat', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ mensagem }) });
    if (!r.ok) throw new Error(await r.text() || 'Erro ' + r.status);
    const leitor = r.body.getReader(), dec = new TextDecoder(); let buf = '';
    for (;;) {
      const { done, value } = await leitor.read(); if (done) break;
      buf += dec.decode(value, { stream: true });
      let i;
      while ((i = buf.indexOf('\n')) >= 0) {
        const linha = buf.slice(0, i).trim(); buf = buf.slice(i + 1); if (!linha) continue;
        const ev = JSON.parse(linha);
        if (ev.type === 'message' && ev.role === 'assistant') {
          if (!resp) resp = add('a', '');
          resp.textContent += ev.content || ''; msgs.scrollTop = msgs.scrollHeight;
        } else if (ev.type === 'tool_use') {
          const f = ev.parameters && (ev.parameters.file_path || ev.parameters.dir_path || ev.parameters.pattern || '');
          estado(true, 'Consultando: ' + (f || ev.tool_name)); resp = null;
        } else if (ev.type === 'fim') {
          interrompido = false;
          if (ev.code) add('e', ev.erro || 'O Gemini terminou com erro (código ' + ev.code + ').');
        }
      }
    }
  } catch (e) { add('e', e.message); interrompido = false; }
  if (interrompido) add('info', 'Execução interrompida.');
  estado(false); txt.focus();
}
fetch('/api/info').then(r => r.json()).then(i => { $('autor').textContent = i.usuario; });
