import { useState, useEffect, useRef, type ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '@/hooks/useAuth'
import {
  ArrowRight, Sparkles, MessageSquare, GraduationCap, BarChart3,
  Link2, Bot, Zap, Check, Globe, Shield, LayoutGrid,
  Menu, X, Code2, ChevronDown
} from 'lucide-react'

// --- Scroll Reveal Hook ---
function useReveal(threshold = 0.15) {
  const ref = useRef<HTMLDivElement>(null)
  const [visible, setVisible] = useState(false)
  useEffect(() => {
    const el = ref.current
    if (!el) return
    const obs = new IntersectionObserver(([e]) => { if (e?.isIntersecting) { setVisible(true); obs.disconnect() } }, { threshold })
    obs.observe(el)
    return () => obs.disconnect()
  }, [threshold])
  return { ref, className: `transition-all duration-700 ease-out ${visible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-8'}` }
}

function Reveal({ children, className = '', threshold }: { children: ReactNode; className?: string; threshold?: number }) {
  const { ref, className: anim } = useReveal(threshold)
  return <div ref={ref} className={`${anim} ${className}`}>{children}</div>
}

// --- Types for mock playground chat ---
interface ChatMessage {
  id: string
  sender: 'ai' | 'user'
  text: string
  timestamp: string
}

export default function LandingPage() {
  const { user } = useAuth()
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)
  const [isAnnual, setIsAnnual] = useState(false)
  const [openFaq, setOpenFaq] = useState<number | null>(null)

  const [messages, setMessages] = useState<ChatMessage[]>([
    {
      id: '1',
      sender: 'ai',
      text: 'Hi! I am the Noant AI assistant. Click one of the questions below or type yours to see how I handle customer queries using custom trained Q&A.',
      timestamp: 'Just now'
    }
  ])
  const [inputValue, setInputValue] = useState('')
  const [isTyping, setIsTyping] = useState(false)

  const presetQuestions = [
    { label: 'How does human takeover work?', answer: 'If Noant AI encounters a question with low confidence (below 0.65 similarity), it flags the chat as "Escalated". Support agents are instantly notified via WebSockets, and they can take over the conversation immediately from the dashboard.' },
    { label: 'Can I connect my WhatsApp number?', answer: 'Yes! Noant supports a self-hosted OpenWA WhatsApp container. You scan a QR code from the dashboard, and your WhatsApp conversations are instantly routed through Noant in real time.' },
    { label: 'How do I train the AI?', answer: 'You can train Noant in seconds by uploading a CSV spreadsheet containing Q&A pairs, creating manual training categories, or typing custom answers directly in the Noant dashboard.' }
  ]

  const faqs = [
    { q: 'How long does it take to set up?', a: 'Most teams are live in under 15 minutes. Sign up, connect your first channel (Web Widget, WhatsApp, or Telegram), upload your Q&A training data, and you\'re ready to go.' },
    { q: 'Does the AI hallucinate answers?', a: 'No. Noant uses a similarity-based retrieval system — it only answers when confidence exceeds your configured threshold (default 65%). Low-confidence questions are automatically escalated to a human agent.' },
    { q: 'Can I customize the AI persona?', a: 'Absolutely. You set the bot name, greeting message, and brand color. The AI adapts its tone based on your training data and system instructions.' },
    { q: 'What happens when the AI can\'t answer?', a: 'The conversation is automatically escalated. Support agents receive a real-time WebSocket notification and can take over the chat instantly from the dashboard. No customer is left waiting.' },
    { q: 'Is my data secure?', a: 'Yes. All data is encrypted in transit (TLS) and at rest. You can self-host the entire stack including the WhatsApp container. No customer data leaves your infrastructure unless you choose cloud deployment.' },
    { q: 'Can I import training data from a spreadsheet?', a: 'Yes. Upload a CSV file with question-answer pairs and categories. The system parses, validates, and indexes your data for instant AI retrieval.' },
  ]

  const handleSendMessage = (text: string) => {
    if (!text.trim() || isTyping) return

    const userMsg: ChatMessage = {
      id: Date.now().toString(),
      sender: 'user',
      text,
      timestamp: 'Just now'
    }
    setMessages((prev) => [...prev, userMsg])
    setInputValue('')
    setIsTyping(true)

    setTimeout(() => {
      let aiText = "That's a great question! I'm trained on your specific business documentation to answer that instantly. To connect this live AI assistant to your channels, sign up for a free trial."
      
      const matchedPreset = presetQuestions.find(q => q.label.toLowerCase() === text.toLowerCase())
      if (matchedPreset) {
        aiText = matchedPreset.answer
      } else if (text.toLowerCase().includes('price') || text.toLowerCase().includes('cost')) {
        aiText = 'Noant plans start at $0/mo for our Free tier (1 AI Agent, 100 chats/mo). Our Pro plan is $49/mo (unlimited chats, custom WhatsApp/Telegram sync) and Enterprise starts at $149/mo.'
      } else if (text.toLowerCase().includes('demo') || text.toLowerCase().includes('try')) {
        aiText = 'You are playing with it right now! Noant runs a similar custom-trained model on your website widget. Create a free account to customize your bot persona!'
      }

      const aiMsg: ChatMessage = {
        id: (Date.now() + 1).toString(),
        sender: 'ai',
        text: aiText,
        timestamp: 'Just now'
      }
      setMessages((prev) => [...prev, aiMsg])
      setIsTyping(false)
    }, 1500)
  }

  const scrollToId = (id: string) => {
    const el = document.getElementById(id)
    if (el) {
      el.scrollIntoView({ behavior: 'smooth' })
      setMobileMenuOpen(false)
    }
  }

  return (
    <div className="min-h-screen bg-[#000000] text-[#e5e5e5] font-sans antialiased overflow-x-hidden">
      {/* Decorative Background Gradients */}
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-full max-w-7xl h-[600px] overflow-hidden pointer-events-none z-0">
        <div className="absolute top-[-20%] left-[20%] w-[500px] h-[500px] rounded-full bg-sky-500/10 blur-[120px] animate-pulse duration-[8s]" />
        <div className="absolute top-[-10%] right-[20%] w-[450px] h-[450px] rounded-full bg-indigo-500/10 blur-[130px] animate-pulse duration-[12s]" />
      </div>

      {/* Floating Navbar */}
      <header className="fixed top-0 left-0 right-0 z-50 bg-[#0d0d0d]/80 backdrop-blur-md border-b border-[#262626] transition-all duration-300">
        <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
          <div className="flex items-center gap-2.5 cursor-pointer" onClick={() => scrollToId('hero')}>
            <svg className="w-7 h-7 text-sky-400" viewBox="0 0 200 200" fill="none">
              <circle cx="100" cy="100" r="92" stroke="currentColor" strokeWidth="10" strokeDasharray="14 18" strokeLinecap="round" />
              <circle cx="100" cy="100" r="70" fill="currentColor" />
              <circle cx="80" cy="100" r="6" fill="#000000" />
              <circle cx="100" cy="100" r="8" fill="#000000" />
              <circle cx="120" cy="100" r="10" fill="#000000" />
            </svg>
            <span className="text-lg font-bold tracking-widest lowercase text-sky-400">noant</span>
          </div>

          <nav className="hidden md:flex items-center gap-8">
            <button onClick={() => scrollToId('features')} className="text-sm text-zinc-400 hover:text-white transition-colors">Features</button>
            <button onClick={() => scrollToId('how-it-works')} className="text-sm text-zinc-400 hover:text-white transition-colors">How It Works</button>
            <button onClick={() => scrollToId('channels')} className="text-sm text-zinc-400 hover:text-white transition-colors">Channels</button>
            <button onClick={() => scrollToId('playground')} className="text-sm text-zinc-400 hover:text-white transition-colors">Live Demo</button>
            <button onClick={() => scrollToId('pricing')} className="text-sm text-zinc-400 hover:text-white transition-colors">Pricing</button>
          </nav>

          <div className="hidden md:flex items-center gap-4">
            {user ? (
              <Link to="/chats" className="inline-flex items-center gap-1.5 px-4 py-2 rounded-xl text-sm font-semibold bg-sky-500 hover:bg-sky-600 text-white shadow-lg shadow-sky-500/20 active:scale-[0.98] transition-all duration-200">
                Go to Dashboard
                <ArrowRight className="w-4 h-4" />
              </Link>
            ) : (
              <>
                <Link to="/login" className="text-sm font-semibold text-zinc-300 hover:text-white transition-colors px-3 py-2">Sign in</Link>
                <Link to="/signup" className="inline-flex items-center gap-1 px-4 py-2 rounded-xl text-sm font-semibold bg-zinc-900 hover:bg-zinc-800 text-white border border-zinc-800 hover:border-zinc-700 active:scale-[0.98] transition-all duration-200">
                  Get started
                </Link>
              </>
            )}
          </div>

          <button onClick={() => setMobileMenuOpen(!mobileMenuOpen)} className="md:hidden p-2 text-zinc-400 hover:text-white focus:outline-none">
            {mobileMenuOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
          </button>
        </div>

        {mobileMenuOpen && (
          <div className="md:hidden bg-[#0d0d0d] border-b border-[#262626] px-6 py-4 space-y-4">
            <button onClick={() => scrollToId('features')} className="block w-full text-left py-2 text-sm text-zinc-400 hover:text-white transition-colors">Features</button>
            <button onClick={() => scrollToId('how-it-works')} className="block w-full text-left py-2 text-sm text-zinc-400 hover:text-white transition-colors">How It Works</button>
            <button onClick={() => scrollToId('channels')} className="block w-full text-left py-2 text-sm text-zinc-400 hover:text-white transition-colors">Channels</button>
            <button onClick={() => scrollToId('playground')} className="block w-full text-left py-2 text-sm text-zinc-400 hover:text-white transition-colors">Live Demo</button>
            <button onClick={() => scrollToId('pricing')} className="block w-full text-left py-2 text-sm text-zinc-400 hover:text-white transition-colors">Pricing</button>
            <div className="pt-4 border-t border-[#262626] flex flex-col gap-3">
              {user ? (
                <Link to="/chats" className="w-full text-center py-2.5 rounded-xl text-sm font-semibold bg-sky-500 text-white" onClick={() => setMobileMenuOpen(false)}>Go to Dashboard</Link>
              ) : (
                <>
                  <Link to="/login" className="w-full text-center py-2.5 rounded-xl text-sm font-semibold text-zinc-300 hover:text-white transition-colors border border-zinc-800" onClick={() => setMobileMenuOpen(false)}>Sign in</Link>
                  <Link to="/signup" className="w-full text-center py-2.5 rounded-xl text-sm font-semibold bg-zinc-900 text-white border border-zinc-800" onClick={() => setMobileMenuOpen(false)}>Get started</Link>
                </>
              )}
            </div>
          </div>
        )}
      </header>

      {/* Hero Section */}
      <section id="hero" className="relative pt-32 pb-24 md:pt-40 md:pb-32 px-6 overflow-hidden">
        <div className="max-w-4xl mx-auto text-center space-y-8 relative z-10">
          <Reveal>
            <div className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-sky-500/10 border border-sky-500/20 text-sky-400 text-xs font-semibold tracking-wider uppercase">
              <Sparkles className="w-3.5 h-3.5" />
              <span>Omnichannel Support Platform</span>
            </div>
          </Reveal>

          <Reveal>
            <h1 className="text-4xl sm:text-5xl md:text-6xl font-extrabold text-white tracking-tight leading-[1.1]">
              Customer Support, <br className="hidden sm:inline" />
              <span className="bg-gradient-to-r from-sky-400 via-sky-500 to-indigo-500 bg-clip-text text-transparent">
                Autopilot & Human
              </span>{' '}
              in Sync
            </h1>
          </Reveal>

          <Reveal>
            <p className="text-base sm:text-lg md:text-xl text-zinc-400 max-w-2xl mx-auto leading-relaxed font-light">
              Empower your team with a custom-trained AI assistant that resolves queries on WhatsApp, Telegram, and your website widget instantly. When confidence drops, support agents take over seamlessly.
            </p>
          </Reveal>

          <Reveal>
            <div className="flex flex-col sm:flex-row items-center justify-center gap-4 pt-4">
              <Link
                to={user ? "/chats" : "/signup"}
                className="w-full sm:w-auto inline-flex items-center justify-center gap-2 px-6 py-3.5 rounded-xl text-sm font-semibold bg-sky-500 hover:bg-sky-600 text-white shadow-lg shadow-sky-500/20 hover:shadow-sky-500/30 active:scale-[0.98] transition-all duration-200 group"
              >
                {user ? 'Go to Dashboard' : 'Start Free Trial'}
                <ArrowRight className="w-4 h-4 group-hover:translate-x-0.5 transition-transform" />
              </Link>
              <button
                onClick={() => scrollToId('playground')}
                className="w-full sm:w-auto inline-flex items-center justify-center gap-1.5 px-6 py-3.5 rounded-xl text-sm font-semibold bg-zinc-900 hover:bg-zinc-800 text-white border border-zinc-800 hover:border-zinc-700 active:scale-[0.98] transition-all duration-200"
              >
                Test AI Playground
              </button>
            </div>
          </Reveal>

          {/* Trust indicators */}
          <Reveal>
            <div className="flex items-center justify-center gap-6 pt-6 text-zinc-500 text-xs">
              <span className="flex items-center gap-1.5"><Check className="w-3.5 h-3.5 text-sky-400" /> Free forever tier</span>
              <span className="flex items-center gap-1.5"><Check className="w-3.5 h-3.5 text-sky-400" /> No credit card required</span>
              <span className="flex items-center gap-1.5"><Check className="w-3.5 h-3.5 text-sky-400" /> Setup in 15 minutes</span>
            </div>
          </Reveal>
        </div>

        {/* Dashboard Mockup Showcase */}
        <Reveal>
          <div className="max-w-5xl mx-auto mt-16 md:mt-24 px-4 sm:px-6 relative z-10">
            <div className="absolute inset-0 bg-gradient-to-tr from-sky-500/10 via-indigo-500/10 to-transparent rounded-2xl blur-xl opacity-60 pointer-events-none" />
            <div className="relative bg-[#0d0d0d] border border-[#262626] rounded-2xl overflow-hidden shadow-2xl">
              <div className="h-11 bg-[#0d0d0d] border-b border-[#262626] px-4 flex items-center justify-between">
                <div className="flex gap-2">
                  <span className="w-3 h-3 rounded-full bg-red-500/30" />
                  <span className="w-3 h-3 rounded-full bg-yellow-500/30" />
                  <span className="w-3 h-3 rounded-full bg-green-500/30" />
                </div>
                <span className="text-[11px] text-zinc-500 font-medium tracking-wide">noant-dashboard.local</span>
                <div className="w-10" />
              </div>
              <div className="grid grid-cols-12 h-[340px] sm:h-[420px] md:h-[480px] bg-[#000000]">
                <div className="hidden sm:block col-span-3 border-r border-[#1a1a1a] p-3 space-y-4">
                  <div className="h-8 bg-zinc-900 rounded-lg animate-pulse" />
                  <div className="space-y-2.5">
                    <div className="h-7 bg-sky-950/20 text-sky-400 rounded-lg flex items-center px-2 text-xs font-semibold gap-2 border border-sky-950">
                      <LayoutGrid className="w-3.5 h-3.5" /> Overview
                    </div>
                    <div className="h-7 hover:bg-zinc-950 text-zinc-500 rounded-lg flex items-center px-2 text-xs gap-2">
                      <MessageSquare className="w-3.5 h-3.5" /> Conversations
                    </div>
                    <div className="h-7 hover:bg-zinc-950 text-zinc-500 rounded-lg flex items-center px-2 text-xs gap-2">
                      <GraduationCap className="w-3.5 h-3.5" /> Teach AI
                    </div>
                    <div className="h-7 hover:bg-zinc-950 text-zinc-500 rounded-lg flex items-center px-2 text-xs gap-2">
                      <Link2 className="w-3.5 h-3.5" /> Integrations
                    </div>
                  </div>
                </div>
                <div className="col-span-12 sm:col-span-9 p-4 md:p-6 space-y-6 overflow-y-auto">
                  <div className="flex justify-between items-center">
                    <div className="space-y-1">
                      <h3 className="text-sm font-bold text-white">Live Operations Overview</h3>
                      <p className="text-[10px] text-zinc-500">Real-time resolution rates and connected channel status.</p>
                    </div>
                    <span className="inline-flex items-center gap-1 text-[9px] text-green-400 bg-green-950/30 border border-green-900/50 px-2 py-0.5 rounded-full font-semibold">
                      <span className="w-1.5 h-1.5 rounded-full bg-green-400 animate-ping" /> Live Syncing
                    </span>
                  </div>
                  <div className="grid grid-cols-3 gap-3 md:gap-4">
                    <div className="bg-[#0d0d0d] border border-[#1a1a1a] rounded-xl p-3.5 space-y-1">
                      <span className="text-[9px] text-zinc-500 uppercase font-semibold">AI Resolution</span>
                      <div className="text-base sm:text-2xl font-extrabold text-sky-400">84.2%</div>
                    </div>
                    <div className="bg-[#0d0d0d] border border-[#1a1a1a] rounded-xl p-3.5 space-y-1">
                      <span className="text-[9px] text-zinc-500 uppercase font-semibold">Open Chats</span>
                      <div className="text-base sm:text-2xl font-extrabold text-white">12</div>
                    </div>
                    <div className="bg-[#0d0d0d] border border-[#1a1a1a] rounded-xl p-3.5 space-y-1">
                      <span className="text-[9px] text-zinc-500 uppercase font-semibold">Handoffs</span>
                      <div className="text-base sm:text-2xl font-extrabold text-indigo-400">3</div>
                    </div>
                  </div>
                  <div className="border border-[#1a1a1a] rounded-xl p-4 bg-[#070707] space-y-3">
                    <div className="flex justify-between items-center text-[10px] text-zinc-500 border-b border-[#1a1a1a] pb-2">
                      <span className="font-semibold text-zinc-400">WhatsApp Session — Active Chat</span>
                      <span>+44 7700 900123</span>
                    </div>
                    <div className="space-y-2.5">
                      <div className="flex gap-2">
                        <div className="w-5 h-5 rounded-full bg-zinc-800 text-white flex items-center justify-center text-[9px] font-bold">C</div>
                        <div className="bg-[#111111] text-zinc-300 text-[11px] p-2.5 rounded-r-xl rounded-bl-xl max-w-[80%] border border-[#1e1e2e]">
                          Hi, do you offer weekend shipping?
                        </div>
                      </div>
                      <div className="flex gap-2 justify-end">
                        <div className="bg-sky-950/40 text-sky-300 text-[11px] p-2.5 rounded-l-xl rounded-br-xl max-w-[80%] border border-sky-900/50">
                          Hello! Yes, shipping is processed 7 days a week. Orders placed before 2 PM on weekends are shipped the same day. 🚚
                        </div>
                        <div className="w-5 h-5 rounded-full bg-sky-500 text-white flex items-center justify-center text-[9px] font-bold">A</div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </Reveal>
      </section>

      {/* How It Works Section */}
      <section id="how-it-works" className="py-24 px-6 border-t border-[#1a1a1a] bg-[#030303] relative">
        <div className="absolute top-[20%] left-[5%] w-[200px] h-[200px] rounded-full bg-indigo-500/5 blur-[80px] pointer-events-none" />
        <div className="max-w-7xl mx-auto space-y-16">
          <Reveal>
            <div className="max-w-2xl mx-auto text-center space-y-4">
              <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-indigo-500/10 border border-indigo-500/20 text-indigo-400 text-[10px] font-bold uppercase tracking-wider">
                <Zap className="w-3.5 h-3.5" />
                <span>3 Steps to Launch</span>
              </div>
              <h2 className="text-2xl sm:text-3xl font-extrabold text-white">
                From Zero to Live AI Support in Minutes
              </h2>
              <p className="text-zinc-400 text-sm sm:text-base leading-relaxed">
                No complex setup. No ML expertise needed. Connect your channels, train your AI, and watch it resolve queries automatically.
              </p>
            </div>
          </Reveal>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 max-w-5xl mx-auto relative">
            {/* Connector line (desktop only) */}
            <div className="hidden md:block absolute top-16 left-[20%] right-[20%] h-px bg-gradient-to-r from-sky-500/30 via-indigo-500/30 to-purple-500/30" />

            <Reveal threshold={0.2}>
              <div className="relative bg-[#0d0d0d] border border-[#262626] hover:border-sky-500/30 rounded-2xl p-8 text-center space-y-4 transition-all duration-300">
                <div className="w-12 h-12 rounded-2xl bg-sky-500/10 border border-sky-500/20 flex items-center justify-center text-sky-400 text-lg font-extrabold mx-auto relative z-10">
                  1
                </div>
                <h3 className="text-lg font-bold text-white">Connect Channels</h3>
                <p className="text-sm text-zinc-400 leading-relaxed">
                  Link your WhatsApp, Telegram, or embed the web widget with a single script tag. No code changes required.
                </p>
                <div className="flex items-center justify-center gap-2 pt-2">
                  <span className="text-[10px] px-2 py-0.5 rounded-full bg-green-500/10 text-green-400 border border-green-900/50">WhatsApp</span>
                  <span className="text-[10px] px-2 py-0.5 rounded-full bg-sky-500/10 text-sky-400 border border-sky-900/50">Telegram</span>
                  <span className="text-[10px] px-2 py-0.5 rounded-full bg-purple-500/10 text-purple-400 border border-purple-900/50">Web Widget</span>
                </div>
              </div>
            </Reveal>

            <Reveal threshold={0.2}>
              <div className="relative bg-[#0d0d0d] border border-[#262626] hover:border-indigo-500/30 rounded-2xl p-8 text-center space-y-4 transition-all duration-300">
                <div className="w-12 h-12 rounded-2xl bg-indigo-500/10 border border-indigo-500/20 flex items-center justify-center text-indigo-400 text-lg font-extrabold mx-auto relative z-10">
                  2
                </div>
                <h3 className="text-lg font-bold text-white">Train Your AI</h3>
                <p className="text-sm text-zinc-400 leading-relaxed">
                  Upload a CSV of Q&A pairs, add training categories manually, or paste custom answers. Your AI learns in seconds.
                </p>
                <div className="flex items-center justify-center gap-2 pt-2">
                  <span className="text-[10px] px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-900/50">CSV Import</span>
                  <span className="text-[10px] px-2 py-0.5 rounded-full bg-amber-500/10 text-amber-400 border border-amber-900/50">Manual Q&A</span>
                </div>
              </div>
            </Reveal>

            <Reveal threshold={0.2}>
              <div className="relative bg-[#0d0d0d] border border-[#262626] hover:border-purple-500/30 rounded-2xl p-8 text-center space-y-4 transition-all duration-300">
                <div className="w-12 h-12 rounded-2xl bg-purple-500/10 border border-purple-500/20 flex items-center justify-center text-purple-400 text-lg font-extrabold mx-auto relative z-10">
                  3
                </div>
                <h3 className="text-lg font-bold text-white">Go Live</h3>
                <p className="text-sm text-zinc-400 leading-relaxed">
                  Your AI starts resolving queries instantly. Low-confidence questions are automatically escalated to human agents via real-time notifications.
                </p>
                <div className="flex items-center justify-center gap-2 pt-2">
                  <span className="text-[10px] px-2 py-0.5 rounded-full bg-rose-500/10 text-rose-400 border border-rose-900/50">Auto-Escalation</span>
                  <span className="text-[10px] px-2 py-0.5 rounded-full bg-sky-500/10 text-sky-400 border border-sky-900/50">Real-time</span>
                </div>
              </div>
            </Reveal>
          </div>
        </div>
      </section>

      {/* Feature Showcase Grid */}
      <section id="features" className="py-24 px-6 border-t border-[#1a1a1a]">
        <div className="max-w-7xl mx-auto space-y-16">
          <Reveal>
            <div className="max-w-2xl mx-auto text-center space-y-4">
              <h2 className="text-2xl sm:text-3xl font-extrabold text-white">
                SaaS Automation Meets Human Control
              </h2>
              <p className="text-zinc-400 text-sm sm:text-base leading-relaxed">
                Noant bridges the gap between hyper-fast autonomous AI responses and the high-fidelity touch of support agents.
              </p>
            </div>
          </Reveal>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 sm:gap-8">
            {[
              { icon: Bot, color: 'sky', title: 'Autonomous AI Autopilot', desc: 'Automatically resolves common queries, checks product inventory, and coordinates details with customers instantly using custom-trained models.' },
              { icon: Sparkles, color: 'indigo', title: 'Human-in-the-Loop Handoff', desc: 'When the AI encounters ambiguous questions or low similarity matches, the conversation escalates instantly so support agents can take over.' },
              { icon: GraduationCap, color: 'emerald', title: 'No-Hallucination Training', desc: 'Upload CSV sheets or add training categories. Noant AI only answers questions matched with high similarity, guaranteeing accurate answers.' },
              { icon: Zap, color: 'purple', title: 'Sub-10ms WebSockets', desc: 'Real-time connection syncing, typing states, and status badges are broadcasted concurrently with thread safety using our custom sync engine.' },
              { icon: Shield, color: 'amber', title: 'Robust Offline Resiliency', desc: 'A custom hook tracks connection status dynamically, locks crucial input components during disconnects, and restores seamlessly on connection.' },
              { icon: BarChart3, color: 'rose', title: 'Lead Tracking & Insights', desc: 'Auto-classifies high-intent conversations, tags user contact profiles, and captures pipeline leads directly into your dashboard.' },
            ].map((f, i) => (
              <Reveal key={i} threshold={0.1}>
                <div className={`bg-[#0d0d0d] border border-[#262626] hover:border-${f.color}-500/50 rounded-2xl p-6 transition-all duration-300 group h-full`}>
                  <div className={`w-10 h-10 rounded-xl bg-${f.color}-500/10 flex items-center justify-center text-${f.color}-400 mb-6 group-hover:scale-110 transition-transform`}>
                    <f.icon className="w-5 h-5" />
                  </div>
                  <h3 className="text-lg font-bold text-white mb-2">{f.title}</h3>
                  <p className="text-sm text-zinc-400 leading-relaxed">{f.desc}</p>
                </div>
              </Reveal>
            ))}
          </div>
        </div>
      </section>

      {/* Connected Channels Showcase */}
      <section id="channels" className="py-24 px-6 border-t border-[#1a1a1a] relative overflow-hidden">
        <div className="absolute top-[30%] right-[-10%] w-[300px] h-[300px] rounded-full bg-indigo-500/5 blur-[90px] pointer-events-none" />
        <div className="max-w-7xl mx-auto grid grid-cols-1 lg:grid-cols-12 gap-12 items-center">
          <Reveal className="lg:col-span-5 space-y-6">
            <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-indigo-500/10 border border-indigo-500/20 text-indigo-400 text-[10px] font-bold uppercase tracking-wider">
              <Globe className="w-3.5 h-3.5" />
              <span>Omnichannel Support</span>
            </div>
            <h2 className="text-2xl sm:text-3xl font-extrabold text-white leading-tight">
              Connect Your Support Wherever Customers Are
            </h2>
            <p className="text-zinc-400 text-sm sm:text-base leading-relaxed">
              Noant centralizes messages from multiple channels into a single inbox, allowing your agents to manage multiple platforms concurrently.
            </p>
            <ul className="space-y-3.5 pt-2">
              {[
                'Self-hosted OpenWA WhatsApp container with cookie logouts.',
                'Telegram Bot integration with webhook-driven chat handlers.',
                'Embeddable Web Widget with dynamic configuration and styling.',
              ].map((text, i) => (
                <li key={i} className="flex items-start gap-3">
                  <span className="w-5 h-5 rounded-full bg-sky-500/10 text-sky-400 flex items-center justify-center shrink-0 mt-0.5"><Check className="w-3.5 h-3.5" /></span>
                  <span className="text-zinc-300 text-sm">{text}</span>
                </li>
              ))}
            </ul>
          </Reveal>

          <div className="lg:col-span-7 grid grid-cols-1 sm:grid-cols-2 gap-4">
            {[
              { icon: Link2, color: 'green', badge: 'Active', title: 'WhatsApp Channel', desc: 'Fully synchronized phone pairing, contact avatars, display names, and direct check validations.' },
              { icon: Globe, color: 'sky', badge: 'Configured', title: 'Telegram Bot', desc: 'Connects to your bot token, processes message queues, and broadcasts custom AI responses directly.' },
              { icon: Code2, color: 'purple', badge: 'Embeddable', title: 'Web Chat Widget', desc: 'Embed a highly stylized glassmorphism chat widget onto your website with single line JS script.' },
              { icon: Link2, color: 'zinc', badge: 'Coming Soon', title: 'Social Integrations', desc: 'Direct messaging synchronization for Facebook Page and Instagram Business accounts is under development.' },
            ].map((ch, i) => (
              <Reveal key={i} threshold={0.1}>
                <div className="bg-[#0d0d0d] border border-[#262626] rounded-2xl p-6 flex flex-col justify-between h-[200px] hover:-translate-y-1 duration-350">
                  <div className="flex justify-between items-start">
                    <div className={`w-10 h-10 rounded-xl bg-${ch.color}-500/10 text-${ch.color}-400 flex items-center justify-center`}><ch.icon className="w-5 h-5" /></div>
                    <span className={`text-[10px] text-${ch.color}-400 bg-${ch.color}-950/30 border border-${ch.color}-900/50 px-2 py-0.5 rounded-full font-semibold`}>{ch.badge}</span>
                  </div>
                  <div>
                    <h3 className="font-bold text-white text-base mb-1">{ch.title}</h3>
                    <p className="text-xs text-zinc-400 leading-relaxed">{ch.desc}</p>
                  </div>
                </div>
              </Reveal>
            ))}
          </div>
        </div>
      </section>

      {/* Interactive AI Playground */}
      <section id="playground" className="py-24 px-6 border-t border-[#1a1a1a] bg-[#030303] relative">
        <div className="absolute top-[10%] left-[10%] w-[250px] h-[250px] rounded-full bg-sky-500/5 blur-[80px] pointer-events-none" />
        <div className="max-w-4xl mx-auto space-y-12">
          <Reveal>
            <div className="text-center space-y-4 max-w-2xl mx-auto">
              <h2 className="text-2xl sm:text-3xl font-extrabold text-white">
                Try the Interactive AI Playground
              </h2>
              <p className="text-zinc-400 text-sm sm:text-base leading-relaxed">
                Test Noant's AI in real-time. Choose a preset training question below, or type your own question to see how Noant responds from its custom knowledge base.
              </p>
            </div>
          </Reveal>

          <Reveal>
            <div className="relative max-w-2xl mx-auto bg-[#0d0d0d] border border-[#262626] rounded-2xl overflow-hidden shadow-2xl">
              <div className="bg-[#0f0f0f] border-b border-[#262626] px-5 py-4 flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="w-8 h-8 rounded-full bg-sky-500/10 text-sky-400 flex items-center justify-center">
                    <Bot className="w-4 h-4" />
                  </div>
                  <div>
                    <h3 className="text-xs font-bold text-white flex items-center gap-1.5">
                      Noant Assistant
                      <span className="w-1.5 h-1.5 rounded-full bg-green-400 animate-ping" />
                    </h3>
                    <p className="text-[10px] text-zinc-500">Autonomous Assistant Demo</p>
                  </div>
                </div>
              </div>

              <div className="h-[280px] sm:h-[320px] p-5 overflow-y-auto space-y-4">
                {messages.map((msg) => (
                  <div key={msg.id} className={`flex gap-3 ${msg.sender === 'user' ? 'justify-end' : ''}`}>
                    {msg.sender === 'ai' && (
                      <div className="w-6 h-6 rounded-full bg-sky-500/10 text-sky-400 flex items-center justify-center text-[10px] font-bold shrink-0 mt-0.5">N</div>
                    )}
                    <div className={`p-3 rounded-2xl text-[12px] leading-relaxed max-w-[80%] border ${
                      msg.sender === 'user'
                        ? 'bg-sky-500 text-white border-sky-400 rounded-tr-none'
                        : 'bg-[#141414] text-zinc-300 border-[#262626] rounded-tl-none'
                    }`}>
                      {msg.text}
                    </div>
                  </div>
                ))}
                {isTyping && (
                  <div className="flex gap-3">
                    <div className="w-6 h-6 rounded-full bg-sky-500/10 text-sky-400 flex items-center justify-center text-[10px] font-bold shrink-0 mt-0.5">N</div>
                    <div className="bg-[#141414] text-zinc-400 text-[12px] p-3 rounded-2xl rounded-tl-none border border-[#262626] flex items-center gap-1">
                      <span className="w-1.5 h-1.5 rounded-full bg-zinc-600 animate-bounce" style={{ animationDelay: '0ms' }} />
                      <span className="w-1.5 h-1.5 rounded-full bg-zinc-600 animate-bounce" style={{ animationDelay: '150ms' }} />
                      <span className="w-1.5 h-1.5 rounded-full bg-zinc-600 animate-bounce" style={{ animationDelay: '300ms' }} />
                    </div>
                  </div>
                )}
              </div>

              <div className="px-5 py-3 bg-[#0a0a0a] border-t border-[#1d1d1d] flex flex-wrap gap-2">
                <span className="text-[10px] text-zinc-500 flex items-center w-full mb-1">Click a training query to test:</span>
                {presetQuestions.map((q) => (
                  <button
                    key={q.label}
                    onClick={() => handleSendMessage(q.label)}
                    disabled={isTyping}
                    className="text-[11px] bg-zinc-900 hover:bg-zinc-800 text-zinc-300 hover:text-white border border-zinc-800 px-3 py-1.5 rounded-full transition-all duration-200 text-left active:scale-[0.98] disabled:opacity-50"
                  >
                    {q.label}
                  </button>
                ))}
              </div>

              <form
                onSubmit={(e) => { e.preventDefault(); handleSendMessage(inputValue) }}
                className="bg-[#0f0f0f] border-t border-[#262626] p-4 flex gap-3"
              >
                <input
                  type="text"
                  value={inputValue}
                  onChange={(e) => setInputValue(e.target.value)}
                  placeholder="Ask about pricing, manual takeover, or weekend shipping..."
                  disabled={isTyping}
                  className="flex-1 bg-[#141414] border border-[#262626] rounded-xl px-4 py-2.5 text-xs text-zinc-350 placeholder-zinc-600 focus:outline-none focus:border-sky-500 transition-colors disabled:opacity-50"
                />
                <button
                  type="submit"
                  disabled={!inputValue.trim() || isTyping}
                  className="px-4 py-2.5 rounded-xl text-xs font-semibold bg-sky-500 hover:bg-sky-600 text-white active:scale-[0.98] transition-all disabled:opacity-50"
                >
                  Send
                </button>
              </form>
            </div>
          </Reveal>
        </div>
      </section>

      {/* Pricing Section */}
      <section id="pricing" className="py-24 px-6 border-t border-[#1a1a1a]">
        <div className="max-w-7xl mx-auto space-y-16">
          <Reveal>
            <div className="text-center space-y-4 max-w-2xl mx-auto">
              <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-sky-500/10 border border-sky-500/20 text-sky-400 text-[10px] font-bold uppercase tracking-wider">
                <Zap className="w-3.5 h-3.5" />
                <span>Simple, transparent pricing</span>
              </div>
              <h2 className="text-2xl sm:text-3xl font-extrabold text-white">
                Plans that scale with your growth
              </h2>
              <p className="text-zinc-400 text-sm sm:text-base leading-relaxed">
                Start for free and upgrade as your team grows and takes on more conversations.
              </p>
              <div className="flex items-center justify-center gap-3 pt-4">
                <span className={`text-xs ${!isAnnual ? 'text-white font-semibold' : 'text-zinc-500'}`}>Monthly</span>
                <button
                  onClick={() => setIsAnnual(!isAnnual)}
                  className="w-10 h-6 rounded-full bg-zinc-800 p-1 flex items-center relative transition-colors"
                >
                  <div className={`w-4 h-4 rounded-full bg-sky-400 transition-all duration-300 ${isAnnual ? 'translate-x-4' : 'translate-x-0'}`} />
                </button>
                <span className={`text-xs ${isAnnual ? 'text-white font-semibold' : 'text-zinc-500'} flex items-center gap-1.5`}>
                  Annual (Save 20%)
                  <span className="text-[10px] bg-green-500/10 border border-green-500/20 text-green-400 px-1.5 py-0.5 rounded-full font-bold">20% Off</span>
                </span>
              </div>
            </div>
          </Reveal>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 max-w-6xl mx-auto items-stretch">
            {/* Free */}
            <Reveal threshold={0.2}>
              <div className="bg-[#0d0d0d] border border-[#262626] rounded-2xl p-8 flex flex-col justify-between relative hover:border-zinc-700 duration-300 h-full">
                <div className="space-y-6">
                  <div className="space-y-2">
                    <h3 className="text-lg font-bold text-white">Free Sandbox</h3>
                    <p className="text-xs text-zinc-500">Perfect for initial testing and local bot tuning.</p>
                  </div>
                  <div className="flex items-baseline gap-1 text-white">
                    <span className="text-3xl sm:text-4xl font-extrabold">$0</span>
                    <span className="text-xs text-zinc-500">/mo</span>
                  </div>
                  <ul className="space-y-3 pt-4 border-t border-[#1a1a1a]">
                    <li className="flex items-center gap-2.5 text-xs text-zinc-400"><Check className="w-4 h-4 text-sky-400 shrink-0" /> 100 AI chats / month</li>
                    <li className="flex items-center gap-2.5 text-xs text-zinc-400"><Check className="w-4 h-4 text-sky-400 shrink-0" /> 1 Connected Web Widget</li>
                    <li className="flex items-center gap-2.5 text-xs text-zinc-400"><Check className="w-4 h-4 text-sky-400 shrink-0" /> Manual Q&A FAQ training</li>
                    <li className="flex items-center gap-2.5 text-xs text-zinc-600"><X className="w-4 h-4 text-zinc-700 shrink-0" /> No WhatsApp integration</li>
                    <li className="flex items-center gap-2.5 text-xs text-zinc-600"><X className="w-4 h-4 text-zinc-700 shrink-0" /> No Telegram integration</li>
                  </ul>
                </div>
                <Link to="/signup" className="mt-8 w-full text-center py-2.5 rounded-xl text-xs font-semibold bg-zinc-900 border border-zinc-800 text-white hover:bg-zinc-800 transition-colors">
                  Get started for free
                </Link>
              </div>
            </Reveal>

            {/* Pro */}
            <Reveal threshold={0.2}>
              <div className="bg-[#0d0d0d] border-2 border-sky-500 rounded-2xl p-8 flex flex-col justify-between relative hover:border-sky-400 duration-300 h-full">
                <span className="absolute top-0 right-6 -translate-y-1/2 bg-sky-500 text-white text-[9px] font-bold uppercase tracking-wider px-2.5 py-1 rounded-full">Most Popular</span>
                <div className="space-y-6">
                  <div className="space-y-2">
                    <h3 className="text-lg font-bold text-white">Noant Pro</h3>
                    <p className="text-xs text-zinc-500">For active businesses running omnichannel customer support.</p>
                  </div>
                  <div className="flex items-baseline gap-1 text-white">
                    <span className="text-3xl sm:text-4xl font-extrabold">${isAnnual ? '39' : '49'}</span>
                    <span className="text-xs text-zinc-500">/mo</span>
                  </div>
                  <ul className="space-y-3 pt-4 border-t border-[#1a1a1a]">
                    <li className="flex items-center gap-2.5 text-xs text-zinc-300"><Check className="w-4 h-4 text-sky-400 shrink-0" /> Unlimited AI chats</li>
                    <li className="flex items-center gap-2.5 text-xs text-zinc-300"><Check className="w-4 h-4 text-sky-400 shrink-0" /> 1 WhatsApp Session (OpenWA)</li>
                    <li className="flex items-center gap-2.5 text-xs text-zinc-300"><Check className="w-4 h-4 text-sky-400 shrink-0" /> 1 Telegram Support Bot</li>
                    <li className="flex items-center gap-2.5 text-xs text-zinc-300"><Check className="w-4 h-4 text-sky-400 shrink-0" /> Spreadsheet CSV training imports</li>
                    <li className="flex items-center gap-2.5 text-xs text-zinc-300"><Check className="w-4 h-4 text-sky-400 shrink-0" /> Up to 3 team members</li>
                  </ul>
                </div>
                <Link to="/signup" className="mt-8 w-full text-center py-2.5 rounded-xl text-xs font-semibold bg-sky-500 text-white hover:bg-sky-600 shadow-lg shadow-sky-500/20 transition-colors">
                  Start Pro Trial
                </Link>
              </div>
            </Reveal>

            {/* Enterprise */}
            <Reveal threshold={0.2}>
              <div className="bg-[#0d0d0d] border border-[#262626] rounded-2xl p-8 flex flex-col justify-between relative hover:border-zinc-700 duration-300 h-full">
                <div className="space-y-6">
                  <div className="space-y-2">
                    <h3 className="text-lg font-bold text-white">Noant Enterprise</h3>
                    <p className="text-xs text-zinc-500">For organizations requiring tailored scaling and custom models.</p>
                  </div>
                  <div className="flex items-baseline gap-1 text-white">
                    <span className="text-3xl sm:text-4xl font-extrabold">${isAnnual ? '119' : '149'}</span>
                    <span className="text-xs text-zinc-500">/mo</span>
                  </div>
                  <ul className="space-y-3 pt-4 border-t border-[#1a1a1a]">
                    <li className="flex items-center gap-2.5 text-xs text-zinc-400"><Check className="w-4 h-4 text-sky-400 shrink-0" /> Multiple WhatsApp & Telegram lines</li>
                    <li className="flex items-center gap-2.5 text-xs text-zinc-400"><Check className="w-4 h-4 text-sky-400 shrink-0" /> Unlimited team members</li>
                    <li className="flex items-center gap-2.5 text-xs text-zinc-400"><Check className="w-4 h-4 text-sky-400 shrink-0" /> Custom LLM intent models (Groq)</li>
                    <li className="flex items-center gap-2.5 text-xs text-zinc-400"><Check className="w-4 h-4 text-sky-400 shrink-0" /> Dedicated priority support agent</li>
                    <li className="flex items-center gap-2.5 text-xs text-zinc-400"><Check className="w-4 h-4 text-sky-400 shrink-0" /> API integration endpoints</li>
                  </ul>
                </div>
                <Link to="/signup" className="mt-8 w-full text-center py-2.5 rounded-xl text-xs font-semibold bg-zinc-900 border border-zinc-800 text-white hover:bg-zinc-800 transition-colors">
                  Contact sales
                </Link>
              </div>
            </Reveal>
          </div>
        </div>
      </section>

      {/* FAQ Section */}
      <section id="faq" className="py-24 px-6 border-t border-[#1a1a1a] bg-[#030303]">
        <div className="max-w-3xl mx-auto space-y-12">
          <Reveal>
            <div className="text-center space-y-4">
              <h2 className="text-2xl sm:text-3xl font-extrabold text-white">
                Frequently Asked Questions
              </h2>
              <p className="text-zinc-400 text-sm sm:text-base leading-relaxed">
                Everything you need to know about getting started with Noant.
              </p>
            </div>
          </Reveal>

          <Reveal>
            <div className="space-y-3">
              {faqs.map((faq, i) => (
                <div key={i} className="bg-[#0d0d0d] border border-[#262626] rounded-xl overflow-hidden transition-colors hover:border-zinc-700">
                  <button
                    onClick={() => setOpenFaq(openFaq === i ? null : i)}
                    className="w-full flex items-center justify-between px-5 py-4 text-left"
                  >
                    <span className="text-sm font-semibold text-white pr-4">{faq.q}</span>
                    <ChevronDown className={`w-4 h-4 text-zinc-500 shrink-0 transition-transform duration-200 ${openFaq === i ? 'rotate-180' : ''}`} />
                  </button>
                  {openFaq === i && (
                    <div className="px-5 pb-4">
                      <p className="text-sm text-zinc-400 leading-relaxed">{faq.a}</p>
                    </div>
                  )}
                </div>
              ))}
            </div>
          </Reveal>
        </div>
      </section>

      {/* Final CTA */}
      <section className="py-24 px-6 border-t border-[#1a1a1a] relative overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-br from-sky-500/5 via-transparent to-indigo-500/5 pointer-events-none" />
        <Reveal>
          <div className="max-w-3xl mx-auto text-center space-y-8 relative z-10">
            <h2 className="text-3xl sm:text-4xl font-extrabold text-white leading-tight">
              Ready to put your customer support on autopilot?
            </h2>
            <p className="text-zinc-400 text-base sm:text-lg leading-relaxed max-w-xl mx-auto">
              Join teams that resolve 84% of queries with AI — and hand off the rest to humans in seconds.
            </p>
            <div className="flex flex-col sm:flex-row items-center justify-center gap-4 pt-2">
              <Link
                to={user ? "/chats" : "/signup"}
                className="w-full sm:w-auto inline-flex items-center justify-center gap-2 px-8 py-4 rounded-xl text-sm font-semibold bg-sky-500 hover:bg-sky-600 text-white shadow-lg shadow-sky-500/20 hover:shadow-sky-500/30 active:scale-[0.98] transition-all duration-200 group"
              >
                {user ? 'Go to Dashboard' : 'Start Free Trial — No Card Needed'}
                <ArrowRight className="w-4 h-4 group-hover:translate-x-0.5 transition-transform" />
              </Link>
            </div>
            <p className="text-zinc-500 text-xs">Free forever on the Starter plan. Upgrade anytime.</p>
          </div>
        </Reveal>
      </section>

      {/* Footer */}
      <footer className="bg-[#070707] border-t border-[#1a1a1a] py-16 px-6">
        <div className="max-w-7xl mx-auto">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-12 mb-12">
            {/* Brand */}
            <div className="md:col-span-1 space-y-4">
              <div className="flex items-center gap-2.5 text-sky-400">
                <svg className="w-5 h-5 shrink-0" viewBox="0 0 200 200" fill="none">
                  <circle cx="100" cy="100" r="92" stroke="currentColor" strokeWidth="10" strokeDasharray="14 18" strokeLinecap="round" />
                  <circle cx="100" cy="100" r="70" fill="currentColor" />
                </svg>
                <span className="text-sm font-bold tracking-widest lowercase">noant</span>
              </div>
              <p className="text-xs text-zinc-500 leading-relaxed">
                AI-powered omnichannel customer support platform with human-in-the-loop escalation.
              </p>
            </div>

            {/* Product */}
            <div className="space-y-3">
              <h4 className="text-xs font-semibold text-white uppercase tracking-wider">Product</h4>
              <ul className="space-y-2">
                <li><button onClick={() => scrollToId('features')} className="text-xs text-zinc-500 hover:text-white transition-colors">Features</button></li>
                <li><button onClick={() => scrollToId('pricing')} className="text-xs text-zinc-500 hover:text-white transition-colors">Pricing</button></li>
                <li><button onClick={() => scrollToId('channels')} className="text-xs text-zinc-500 hover:text-white transition-colors">Integrations</button></li>
                <li><button onClick={() => scrollToId('playground')} className="text-xs text-zinc-500 hover:text-white transition-colors">Live Demo</button></li>
              </ul>
            </div>

            {/* Resources */}
            <div className="space-y-3">
              <h4 className="text-xs font-semibold text-white uppercase tracking-wider">Resources</h4>
              <ul className="space-y-2">
                <li><a href="/docs" className="text-xs text-zinc-500 hover:text-white transition-colors">Documentation</a></li>
                <li><a href="/docs#api" className="text-xs text-zinc-500 hover:text-white transition-colors">API Reference</a></li>
                <li><a href="/docs#database" className="text-xs text-zinc-500 hover:text-white transition-colors">Database Schema</a></li>
                <li><a href="/changelog" className="text-xs text-zinc-500 hover:text-white transition-colors">Changelog</a></li>
              </ul>
            </div>

            {/* Legal */}
            <div className="space-y-3">
              <h4 className="text-xs font-semibold text-white uppercase tracking-wider">Legal</h4>
              <ul className="space-y-2">
                <li><a href="/privacy" className="text-xs text-zinc-500 hover:text-white transition-colors">Privacy Policy</a></li>
                <li><a href="/terms" className="text-xs text-zinc-500 hover:text-white transition-colors">Terms of Service</a></li>
                <li><a href="/security" className="text-xs text-zinc-500 hover:text-white transition-colors">Security</a></li>
                <li><a href="/license" className="text-xs text-zinc-500 hover:text-white transition-colors">License (MIT)</a></li>
              </ul>
            </div>
          </div>

          <div className="border-t border-[#1a1a1a] pt-8 flex flex-col md:flex-row items-center justify-between gap-4">
            <p className="text-[11px] text-zinc-600">&copy; {new Date().getFullYear()} Noant. All rights reserved.</p>
            <div className="flex items-center gap-4">
              <a href="https://github.com/noant-enterprise/noant" target="_blank" rel="noopener noreferrer" className="text-[11px] text-zinc-600 hover:text-white transition-colors">GitHub</a>
              <a href="mailto:support@noant.dev" className="text-[11px] text-zinc-600 hover:text-white transition-colors">Contact</a>
            </div>
          </div>
        </div>
      </footer>
    </div>
  )
}
