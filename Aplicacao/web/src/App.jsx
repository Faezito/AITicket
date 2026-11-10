import { useState, useRef, useEffect } from 'react'

const post = (url, body, signal) =>
  fetch(url, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body), signal })

const texto = (v) => (typeof v === 'string' ? v : JSON.stringify(v))

function Card({ m, aprovar, rejeitar }) {
  return (
    <div className="card">
      <h3>Proposta de chamado</h3>
      {Object.entries(m.ticket).map(([k, v]) => <p key={k}><b>{k}:</b> {texto(v)}</p>)}
      {m.estado ? <p className="info">{m.estado}</p> : (
        <div className="acoes">
          <button className="pri" onClick={aprovar}>Aprovar</button>
          <button onClick={rejeitar}>Rejeitar</button>
          <span className="info">ou escreva o ajuste abaixo</span>
        </div>
      )}
    </div>
  )
}

export default function App() {
  const [msgs, setMsgs] = useState([])
  const [txt, setTxt] = useState('')
  const [busy, setBusy] = useState(false)
  const [usuario, setUsuario] = useState('')
  const ctl = useRef(null)
  const fim = useRef(null)

  useEffect(() => { fetch('/api/info').then(r => r.json()).then(i => setUsuario(i.usuario)) }, [])
  useEffect(() => { fim.current?.scrollIntoView() }, [msgs, busy])

  const add = (m) => setMsgs(p => [...p, m])
  const set = (i, patch) => setMsgs(p => p.map((m, k) => (k === i ? { ...m, ...patch } : m)))

  async function enviar() {
    const t = txt.trim()
    if (!t || busy) return
    const user = { role: 'user', content: t }
    const hist = [...msgs, user].filter(m => m.role === 'user' || m.role === 'assistant').map(({ role, content }) => ({ role, content }))
    setMsgs([...msgs, user]); setTxt(''); setBusy(true)
    ctl.current = new AbortController()
    try {
      const r = await post('/api/chat', { mensagens: hist }, ctl.current.signal)
      if (!r.ok) throw new Error(await r.text())
      const d = await r.json()
      add({ role: 'assistant', content: d.mensagem, ticket: d.ticket, id: d.id })
    } catch (e) {
      add(e.name === 'AbortError' ? { role: 'info', content: 'Execução interrompida.' } : { role: 'erro', content: e.message })
    }
    setBusy(false)
  }

  async function aprovar(i, m) {
    const r = await post('/api/tickets', { id: m.id, ticket: m.ticket })
    if (!r.ok) return add({ role: 'erro', content: await r.text() })
    set(i, { estado: 'Aprovado e salvo em Historico/' + (await r.json()).arquivo })
  }

  return (
    <>
      <header>
        <b>Chamados</b>
        <div>
          <span className="usr">{usuario}</span>
          <button onClick={() => { ctl.current?.abort(); setMsgs([]) }}>Nova conversa</button>
        </div>
      </header>
      <main>
        <div className="col">
          {!msgs.length && <p className="vazio">Descreva o atendimento para começar.</p>}
          {msgs.map((m, i) => m.ticket
            ? <Card key={i} m={m} aprovar={() => aprovar(i, m)} rejeitar={() => set(i, { estado: 'Rejeitado' })} />
            : <div key={i} className={'m ' + m.role}>{m.content}</div>)}
          {busy && <p className="info">A IA está processando…</p>}
          <div ref={fim} />
        </div>
      </main>
      <footer>
        <div className="col linha">
          <textarea rows={1} value={txt} placeholder="Descreva o atendimento (Shift+Enter para nova linha)"
            onChange={e => setTxt(e.target.value)}
            onKeyDown={e => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); enviar() } }} />
          {busy ? <button onClick={() => ctl.current?.abort()}>Parar</button>
                : <button className="pri" onClick={enviar}>Enviar</button>}
        </div>
      </footer>
    </>
  )
}
