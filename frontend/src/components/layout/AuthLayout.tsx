import { Outlet } from 'react-router-dom'

export function AuthLayout() {
  return (
    <div className="min-h-screen flex">
      {/* Left side - brand (always dark, it's the brand panel) */}
      <div className="hidden lg:flex flex-1 bg-noant-black text-white flex-col justify-between p-12 relative overflow-hidden">
        <div className="relative z-10">
          <div className="flex items-center gap-3 mb-16">
            <svg className="w-10 h-10" viewBox="0 0 100 100" fill="none">
              <circle cx="50" cy="50" r="42" stroke="white" strokeWidth="6" strokeDasharray="8 10" strokeLinecap="round" />
              <circle cx="50" cy="50" r="28" fill="white" />
              <circle cx="40" cy="50" r="3" fill="#0a0a0a" />
              <circle cx="50" cy="50" r="4" fill="#0a0a0a" />
              <circle cx="60" cy="50" r="5" fill="#0a0a0a" />
            </svg>
            <span className="text-2xl font-bold tracking-widest lowercase">noant</span>
          </div>
          <p className="text-3xl font-light leading-relaxed max-w-md">
            "I used to wake up to 47 unread messages. Now I wake up to closed sales."
          </p>
        </div>
        <div className="relative z-10 flex gap-12">
          <div>
            <div className="text-3xl font-bold text-noant-sky">500+</div>
            <div className="text-sm opacity-60 uppercase tracking-widest mt-1">Languages</div>
          </div>
          <div>
            <div className="text-3xl font-bold text-noant-sky">24/7</div>
            <div className="text-sm opacity-60 uppercase tracking-widest mt-1">Always on</div>
          </div>
          <div>
            <div className="text-3xl font-bold text-noant-sky">&lt;2s</div>
            <div className="text-sm opacity-60 uppercase tracking-widest mt-1">Response time</div>
          </div>
        </div>
      </div>

      {/* Right side - form (follows theme) */}
      <div className="flex-1 flex items-center justify-center p-6 lg:p-12 bg-base transition-colors duration-300">
        <div className="w-full max-w-md">
          <Outlet />
        </div>
      </div>
    </div>
  )
}