# Agent and chat (M11)

A research and planning assistant inside moneyd: a backend agent that works
against the household's real data and outside sources, and a chat page that
is a window onto it. This document is the plan; nothing of it exists yet.
It is written to be read by people and by the model that will run the agent.

## What it is for

The main use case: one member of the household wants to start investing, an
amount upfront and a monthly contribution. The owner of the data has rough
ideas about sectors and criteria; the agent does the research, compares
candidates, and turns the result into a plan that the app can show and
simulate. Secondary use: questions about the existing holdings in plain
words, answered from the data ("how is Günter's bond allocation doing?").

Not for: executing anything. moneyd is bookkeeping and planning; there is
no brokerage connection and there will not be one.

## Shape

Three layers in the backend, each independently testable, and a thin client.

### Model client

An OpenAI-compatible chat completions client: `POST /v1/chat/completions`
with `tools`, `tool_choice`, and `stream: true`. Base URL, API key and model
name come from flags or environment (`-agent-url`, `-agent-key`,
`-agent-model`). The first target is a self-hosted model; the same client
reaches any hosted provider that speaks the protocol, so a run that needs
more capability can use a different model without code changes.

Requirements on the model, to be tested before any UI exists:

- Reliable function calling with 15–25 tools in the request and several
  calls per turn, including parallel calls.
- Tool results of a few thousand tokens each without losing the thread;
  a run may hold 30–60 k tokens of context.
- Streaming of both text deltas and tool-call deltas.
- Following a system prompt that forbids inventing figures (see Principles).

### Tool registry

Every tool is a Go function with a name, a one-paragraph description, a JSON
schema for its arguments, and a typed result that is rendered to JSON for
the model and to a card for the UI. The agent sees the registry and nothing
else; there is no raw database or GraphQL access.

Two families.

**Internal** (read the household's data; aggregates by default):

| Tool | Returns |
|---|---|
| `list_persons` | the persons the user can act for, and the active one |
| `holdings` | positions per portfolio: units, cost basis, market value, gain, kind, currency |
| `performance` | time-weighted return and totals for a period, per portfolio or overall |
| `cash_accounts` | balances and currencies |
| `net_worth_history` | net worth against invested capital over a period, as points |
| `realized_gains` | dividends, interest, sales in a period |
| `asset` | one asset: identifiers, listings, latest quote, quote history summary, kind |
| `search_assets` | by name or ISIN among the securities the app knows |
| `transactions` | the transactions of a security or an account in a period (detail; off by default, see Principles) |

**External** (research; every result cached in the database with source and
timestamp):

| Tool | Returns |
|---|---|
| `web_search` | titles, URLs and snippets for a query, via SearXNG or a hosted search API behind one interface |
| `fetch_page` | the readable text of a URL, size-capped, with a per-domain politeness delay; PDFs through pdftotext |
| `dividend_history` | per ISIN: payments, yield, growth, from the configured source |
| `instrument_lookup` | per ISIN or name: what the instrument is, where it trades, current price (ING's instrument search is the first source) |
| `fx_rate` | ECB rate between two currencies on a date (already stored for valuation) |

**Acting** (the only writes, each one visible as a card and reversible in
the UI):

| Tool | Does |
|---|---|
| `add_to_watchlist` | records a candidate with the facts gathered and their sources |
| `create_plan` / `update_plan` | a plan: starting amount, monthly contribution, allocation, horizon (M12) |
| `simulate_plan` | runs the projection for a plan and returns the figures (M12) |

A tool never raises for an empty result; it returns a result that says so.
Errors are returned as results too, so the model can decide to retry or to
tell the user.

### Run executor

A run is one goal, pursued by a loop of model turns and tool calls until the
model stops calling tools or a limit is hit. Every turn, every tool call and
every result is persisted as a step before the next one starts. That gives:

- Resumability: a crash or a restart continues the run from its last step.
- An audit trail: when the report says a number, the step that produced it
  can be shown.
- Progress: the UI renders steps as they are written.

Limits per run: steps, tokens, wall time, and a per-tool budget for
external calls. Exceeding one ends the run with a status that says which.

Runs stream to the client over Server-Sent Events: step started, tool
result, text delta, run finished. A finished run produces a report: the
model's final message, stored, with links from each figure to its step.

### Chat

A thread per person holds messages. A user message may start a run; the run
belongs to the message, and the thread shows the run's steps inline and the
report at the end. Threads, messages and runs are ordinary data: listed,
reopened, deleted. A report can be opened on its own page, which is what
one shows to someone else.

Later, not first: the same tool registry served over MCP, so Claude Code or
another client can drive the agent without the chat page.

## Data model

```
agent_threads     id, person_id, title, created_at
agent_messages    id, thread_id, role (user|assistant), content, run_id?, created_at
agent_runs        id, message_id, goal, status (running|done|failed|limited), model,
                  started_at, finished_at, tokens_in, tokens_out, report (text)
agent_steps       id, run_id, index, kind (turn|tool_call|tool_result|text),
                  tool_name?, arguments (json)?, result (json)?, tokens?, created_at
source_cache      key (tool + arguments hash), tool_name, fetched_at, expires_at,
                  source_url?, result (json)
watchlist         id, person_id, security_id?, isin?, name, added_by_run?, created_at
watchlist_facts   id, watchlist_id, fact (text), source_url?, run_id?, created_at
```

Plans and simulations are M12 and get their own tables then.

## API

- GraphQL: `threads`, `thread(id)`, `createThread`, `sendMessage(threadID,
  text)` which creates the message and starts the run and returns both,
  `cancelRun(id)`, `run(id)` with its steps, `watchlist`, `deleteThread`.
- `GET /agent/runs/{id}/events`: Server-Sent Events for a run, replaying
  stored steps first, then live ones until the run ends. Protected by the
  same session as `/graphql`.

## Principles

- **Numbers come from tools.** The system prompt says it; the report
  renderer enforces it by linking each figure to a step, and a figure with
  no step is shown as unverified. The model composes; it does not recall.
- **Recommendations are constrained.** The agent proposes candidates
  against criteria the user stated, in sectors the user named, and presents
  them as a structured list with the gathered figures, not as persuasive
  prose. It never invents products. The usual caveat that this is not
  financial advice is part of every report.
- **What leaves the house is a choice per tool.** Internal tools return
  aggregates; per-transaction detail is a tool the user enables. With a
  self-hosted model this is moot; with a hosted model it matters.
- **Sources are respected.** External results are cached, fetches are
  rate-limited per domain, and each source's terms are checked before it is
  wired in. Whether DivvyDiary offers an API that permits this use is open;
  the tool interface is designed so the source behind it can change.
- **Everything the agent does is visible and reversible.** Writes are
  limited to the watchlist and plans, both editable in the UI.

## Order of work

1. Model client, tool registry with the internal tools, run executor with
   persistence and SSE, a chat page with streaming steps. Prove the loop on
   questions about the existing holdings.
2. External tools with the cache, the watchlist, reports as pages. This is
   the research milestone and needs the user's sector ideas as test prompts.
3. Plans and simulation (M12) and the tools that drive them, so research
   turns into "put 300 € a month into these four" with a projection.
4. The MCP endpoint over the same registry.

## Open questions

- Which model and hardware will be hosted; this decides how large the tool
  set per request can be and whether a second, stronger model is needed for
  the final report.
- Search: self-hosted SearXNG or a hosted API with a key.
- Dividend data: which source, under which terms.
- How much of the household's detail may go to a hosted model, if one is
  ever used.
