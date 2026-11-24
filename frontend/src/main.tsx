import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './styles.css'
import Keycloak from 'keycloak-js'
import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL
const keycloak = new Keycloak({ url: import.meta.env.VITE_KEYCLOAK_URL, realm: import.meta.env.VITE_KEYCLOAK_REALM, clientId: import.meta.env.VITE_KEYCLOAK_CLIENT_ID })
type Job = { id:string; originalFilename:string; status:string; createdAt:string }

function App() {
  const [ready,setReady]=useState(false), [jobs,setJobs]=useState<Job[]>([]), [error,setError]=useState(''), [file,setFile]=useState<File>()
  const load=async()=>{const r=await fetch(`${apiBaseUrl}/api/v1/videos`,{headers:{Authorization:`Bearer ${keycloak.token}`}});if(!r.ok)throw Error('Não foi possível carregar os vídeos');setJobs((await r.json()).items)}
  useEffect(()=>{keycloak.init({onLoad:'check-sso',pkceMethod:'S256'}).then(async authenticated=>{setReady(true);if(authenticated)try{await load()}catch(e){setError((e as Error).message)}})},[])
  const send=async(e:FormEvent)=>{e.preventDefault();if(!file)return;setError('');try{const token=keycloak.token;const created=await fetch(`${apiBaseUrl}/api/v1/videos/uploads`,{method:'POST',headers:{Authorization:`Bearer ${token}`,'Content-Type':'application/json'},body:JSON.stringify({filename:file.name,contentType:file.type,sizeBytes:file.size})});if(!created.ok)throw Error('Arquivo inválido ou serviço indisponível');const data=await created.json();const put=await fetch(data.uploadUrl,{method:'PUT',headers:{'Content-Type':file.type},body:file});if(!put.ok)throw Error('Falha ao enviar para o armazenamento');const confirmed=await fetch(`${apiBaseUrl}/api/v1/videos/${data.jobId}/confirm`,{method:'POST',headers:{Authorization:`Bearer ${keycloak.token}`}});if(!confirmed.ok)throw Error('Não foi possível confirmar o upload');await load();setFile(undefined)}catch(e){setError((e as Error).message)}}
  if(!ready)return <main><p>Carregando sessão…</p></main>
  if(!keycloak.authenticated)return <main><h1>FIAP Video</h1><p>Envie vídeos com segurança e acompanhe o estado do envio.</p><button onClick={()=>keycloak.login()}>Entrar</button></main>
  return <main><header><h1>FIAP Video</h1><button onClick={()=>keycloak.logout({redirectUri:window.location.origin})}>Sair</button></header><form onSubmit={send}><label>Vídeo <input aria-label="Selecione um vídeo" type="file" accept="video/mp4,video/quicktime,video/x-msvideo,video/x-matroska,video/webm,.mkv" onChange={e=>setFile(e.target.files?.[0])}/></label><button>Enviar</button></form>{error&&<p role="alert">{error}</p>}<h2>Meus vídeos</h2><ul>{jobs.map(j=><li key={j.id}><strong>{j.originalFilename}</strong> — {j.status==='AWAITING_UPLOAD'||j.status==='PENDING'?'Aguardando processamento':j.status}</li>)}</ul></main>
}

createRoot(document.getElementById('root')!).render(<StrictMode><App /></StrictMode>)
