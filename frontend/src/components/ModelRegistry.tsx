import { useCallback, useEffect, useState } from 'react'
import { api, AIModelOption, AIRegistryResponse } from '../lib/api'

const PROVIDERS = ['anthropic', 'openai', 'gemini', 'deepseek']

const emptyForm = {
  provider: 'openai',
  model_id: '',
  label: '',
  base_url: '',
  input_price_per_mtok: '',
  output_price_per_mtok: '',
}

type ActionMsg = { kind: 'ok' | 'err'; text: string }

export default function ModelRegistry() {
  const [data, setData] = useState<AIRegistryResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [busyId, setBusyId] = useState<number | null>(null)
  const [confirmId, setConfirmId] = useState<number | null>(null)
  const [msg, setMsg] = useState<ActionMsg | null>(null)
  const [showAdd, setShowAdd] = useState(false)
  const [adding, setAdding] = useState(false)
  const [form, setForm] = useState({ ...emptyForm })

  const load = useCallback(() => {
    return api
      .aiRegistry()
      .then((d) => setData(d))
      .catch((e) => setMsg({ kind: 'err', text: e.message }))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    load()
  }, [load])

  const registry = data?.registry ?? null
  const enabled = data?.management_enabled ?? false
  const active = registry?.active
  const options = registry?.options ?? []

  const isActive = (o: AIModelOption) =>
    !!active && o.provider === active.active_provider && o.model_id === active.active_model

  async function activate(o: AIModelOption) {
    setBusyId(o.id)
    setMsg(null)
    try {
      await api.activateAIModel(o.id)
      setMsg({ kind: 'ok', text: `Switched the light tier to ${o.label || o.model_id} ✓` })
      await load()
    } catch (e) {
      setMsg({ kind: 'err', text: `Switch refused: ${(e as Error).message}` })
      await load()
    } finally {
      setBusyId(null)
    }
  }

  async function remove(o: AIModelOption) {
    if (confirmId !== o.id) {
      setConfirmId(o.id)
      return
    }
    setConfirmId(null)
    setBusyId(o.id)
    setMsg(null)
    try {
      await api.deleteAIModel(o.id)
      setMsg({ kind: 'ok', text: `Removed ${o.label || o.model_id}` })
      await load()
    } catch (e) {
      setMsg({ kind: 'err', text: (e as Error).message })
    } finally {
      setBusyId(null)
    }
  }

  async function submitAdd(e: React.FormEvent) {
    e.preventDefault()
    setAdding(true)
    setMsg(null)
    try {
      const opt = await api.addAIModel({
        provider: form.provider,
        model_id: form.model_id.trim(),
        label: form.label.trim() || form.model_id.trim(),
        base_url: form.base_url.trim(),
        input_price_per_mtok: parseFloat(form.input_price_per_mtok) || 0,
        output_price_per_mtok: parseFloat(form.output_price_per_mtok) || 0,
        enabled: true,
      })
      if (opt.validated) {
        setMsg({ kind: 'ok', text: `Added ${opt.label || opt.model_id} — validation passed ✓` })
      } else {
        setMsg({
          kind: 'err',
          text: `Added ${opt.label || opt.model_id}, but the validation probe failed: ${opt.validation_error || 'unknown error'}. It can't be activated until it probes green.`,
        })
      }
      setForm({ ...emptyForm })
      setShowAdd(false)
      await load()
    } catch (e) {
      setMsg({ kind: 'err', text: (e as Error).message })
    } finally {
      setAdding(false)
    }
  }

  if (loading && !data) {
    return (
      <div className="bg-[#1E1E28] rounded-lg p-6 border border-[#3A3A48] text-[#F0F0F5]/50 text-sm">
        Loading model registry…
      </div>
    )
  }

  return (
    <div className="bg-[#1E1E28] rounded-lg border border-[#3A3A48]">
      <div className="flex items-center justify-between p-4 border-b border-[#3A3A48]">
        <div>
          <h3 className="text-sm font-medium text-[#F0F0F5]/90">Light-tier Model — Live Switch</h3>
          <p className="text-xs text-[#F0F0F5]/50 mt-0.5">
            Preview / extraction / warming tier. Generation stays on Claude Sonnet.
          </p>
        </div>
        {active && active.active_model ? (
          <div className="text-right">
            <div className="text-[10px] uppercase tracking-wide text-[#F0F0F5]/40">Active</div>
            <div className="text-sm text-[#5CFFD4] tabular-nums">
              {active.active_provider} / {active.active_model}
            </div>
          </div>
        ) : (
          <span className="text-xs text-[#F0F0F5]/40">no active model recorded</span>
        )}
      </div>

      {!enabled && (
        <div className="m-4 px-3 py-2 rounded bg-amber-500/10 border border-amber-500/30 text-amber-300/90 text-xs">
          Live switching is disabled — set <code>API_BASE_URL</code> + <code>API_ID_HEADER</code> +{' '}
          <code>ADMIN_TOKEN</code> on the dashboard to enable. Registry is shown read-only.
        </div>
      )}

      {msg && (
        <div
          className={`m-4 px-3 py-2 rounded text-xs border ${
            msg.kind === 'ok'
              ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-300'
              : 'bg-red-500/10 border-red-500/30 text-red-300'
          }`}
        >
          {msg.text}
        </div>
      )}

      <div className="p-4 overflow-x-auto">
        {options.length === 0 ? (
          <p className="text-sm text-[#F0F0F5]/50 py-4 text-center">
            No models registered yet
            {enabled ? ' — add one below.' : '.'}
          </p>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-[#F0F0F5]/50 border-b border-[#3A3A48]">
                <th className="py-2 pr-4 font-medium">Model</th>
                <th className="py-2 pr-4 font-medium">Provider</th>
                <th className="py-2 pr-4 font-medium text-right">$/Mtok in</th>
                <th className="py-2 pr-4 font-medium text-right">$/Mtok out</th>
                <th className="py-2 pr-4 font-medium">Status</th>
                <th className="py-2 pr-4 font-medium text-right">Actions</th>
              </tr>
            </thead>
            <tbody>
              {options.map((o) => {
                const act = isActive(o)
                const busy = busyId === o.id
                return (
                  <tr key={o.id} className="border-b border-[#3A3A48]/50 hover:bg-[#2A2A36]/40">
                    <td className="py-2.5 pr-4">
                      <div className="flex items-center gap-2">
                        <span className="text-[#F0F0F5]/90">{o.label || o.model_id}</span>
                        {act && (
                          <span className="px-1.5 py-0.5 rounded text-[10px] uppercase tracking-wide bg-[#5CFFD4]/15 text-[#5CFFD4] border border-[#5CFFD4]/30">
                            Active
                          </span>
                        )}
                      </div>
                      <div className="text-xs text-[#F0F0F5]/40 tabular-nums">{o.model_id}</div>
                    </td>
                    <td className="py-2.5 pr-4 text-[#F0F0F5]/70">{o.provider}</td>
                    <td className="py-2.5 pr-4 text-right tabular-nums text-[#F0F0F5]/70">
                      {o.input_price_per_mtok ? `$${o.input_price_per_mtok.toFixed(2)}` : '—'}
                    </td>
                    <td className="py-2.5 pr-4 text-right tabular-nums text-[#F0F0F5]/70">
                      {o.output_price_per_mtok ? `$${o.output_price_per_mtok.toFixed(2)}` : '—'}
                    </td>
                    <td className="py-2.5 pr-4">
                      {o.validated ? (
                        <span className="text-emerald-400 text-xs">✓ validated</span>
                      ) : (
                        <span
                          className="text-red-400/80 text-xs"
                          title={o.validation_error || 'not yet validated'}
                        >
                          ✗ unvalidated
                        </span>
                      )}
                    </td>
                    <td className="py-2.5 pr-4">
                      <div className="flex items-center gap-2 justify-end">
                        <button
                          onClick={() => activate(o)}
                          disabled={!enabled || act || busy}
                          className="px-2.5 py-1 rounded text-xs border border-[#FF6B85]/40 text-[#FF6B85] hover:bg-[#FF6B85]/10 disabled:opacity-30 disabled:cursor-not-allowed"
                        >
                          {busy ? '…' : act ? 'Active' : 'Activate'}
                        </button>
                        <button
                          onClick={() => remove(o)}
                          disabled={!enabled || busy}
                          className="px-2.5 py-1 rounded text-xs border border-[#3A3A48] text-[#F0F0F5]/60 hover:bg-[#3A3A48]/40 disabled:opacity-30 disabled:cursor-not-allowed"
                        >
                          {confirmId === o.id ? 'Confirm?' : 'Delete'}
                        </button>
                      </div>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        )}
      </div>

      {enabled && (
        <div className="px-4 pb-4">
          {!showAdd ? (
            <button
              onClick={() => setShowAdd(true)}
              className="text-xs text-[#FF6B85] hover:underline"
            >
              + Add model
            </button>
          ) : (
            <form
              onSubmit={submitAdd}
              className="mt-2 grid grid-cols-2 md:grid-cols-3 gap-3 p-4 rounded bg-[#15151D] border border-[#3A3A48]"
            >
              <label className="text-xs text-[#F0F0F5]/60">
                Provider
                <select
                  value={form.provider}
                  onChange={(e) => setForm({ ...form, provider: e.target.value })}
                  className="mt-1 w-full bg-[#1E1E28] border border-[#3A3A48] rounded px-2 py-1.5 text-sm text-[#F0F0F5]"
                >
                  {PROVIDERS.map((p) => (
                    <option key={p} value={p}>
                      {p}
                    </option>
                  ))}
                </select>
              </label>
              <label className="text-xs text-[#F0F0F5]/60">
                Model ID
                <input
                  required
                  value={form.model_id}
                  onChange={(e) => setForm({ ...form, model_id: e.target.value })}
                  placeholder="gpt-4o-mini"
                  className="mt-1 w-full bg-[#1E1E28] border border-[#3A3A48] rounded px-2 py-1.5 text-sm text-[#F0F0F5]"
                />
              </label>
              <label className="text-xs text-[#F0F0F5]/60">
                Label
                <input
                  value={form.label}
                  onChange={(e) => setForm({ ...form, label: e.target.value })}
                  placeholder="GPT-4o mini"
                  className="mt-1 w-full bg-[#1E1E28] border border-[#3A3A48] rounded px-2 py-1.5 text-sm text-[#F0F0F5]"
                />
              </label>
              <label className="text-xs text-[#F0F0F5]/60">
                Base URL (optional)
                <input
                  value={form.base_url}
                  onChange={(e) => setForm({ ...form, base_url: e.target.value })}
                  placeholder="provider default"
                  className="mt-1 w-full bg-[#1E1E28] border border-[#3A3A48] rounded px-2 py-1.5 text-sm text-[#F0F0F5]"
                />
              </label>
              <label className="text-xs text-[#F0F0F5]/60">
                $ / Mtok input
                <input
                  type="number"
                  step="0.001"
                  value={form.input_price_per_mtok}
                  onChange={(e) => setForm({ ...form, input_price_per_mtok: e.target.value })}
                  placeholder="0.15"
                  className="mt-1 w-full bg-[#1E1E28] border border-[#3A3A48] rounded px-2 py-1.5 text-sm text-[#F0F0F5]"
                />
              </label>
              <label className="text-xs text-[#F0F0F5]/60">
                $ / Mtok output
                <input
                  type="number"
                  step="0.001"
                  value={form.output_price_per_mtok}
                  onChange={(e) => setForm({ ...form, output_price_per_mtok: e.target.value })}
                  placeholder="0.60"
                  className="mt-1 w-full bg-[#1E1E28] border border-[#3A3A48] rounded px-2 py-1.5 text-sm text-[#F0F0F5]"
                />
              </label>
              <div className="col-span-2 md:col-span-3 flex items-center gap-3 pt-1">
                <button
                  type="submit"
                  disabled={adding}
                  className="px-3 py-1.5 rounded text-sm bg-[#FF6B85]/15 border border-[#FF6B85]/40 text-[#FF6B85] hover:bg-[#FF6B85]/25 disabled:opacity-40"
                >
                  {adding ? 'Validating…' : 'Add & validate'}
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setShowAdd(false)
                    setForm({ ...emptyForm })
                  }}
                  className="text-xs text-[#F0F0F5]/50 hover:text-[#F0F0F5]/80"
                >
                  Cancel
                </button>
                <span className="text-xs text-[#F0F0F5]/40">
                  Adding runs a live validation probe; a model only goes live once it passes.
                </span>
              </div>
            </form>
          )}
        </div>
      )}
    </div>
  )
}
