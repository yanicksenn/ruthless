import React from 'react';
import { useAuth } from '../context/AuthContext';
import { LogIn, UserPlus, ShieldAlert } from 'lucide-react';

export const Login: React.FC = () => {
  const { user, loginWithToken, completeRegistration, isDevelopment, limits } = useAuth();
  const [error, setError] = React.useState<string | null>(null);
  const [chosenName, setChosenName] = React.useState('');

  const minLen = limits?.minUserNameLength ?? 2;
  const maxLen = limits?.maxUserNameLength ?? 32;
  const isInvalidLength = chosenName.length > 0 && (chosenName.length < minLen || chosenName.length > maxLen);

  const handleGoogleLogin = () => {
    const baseUrl = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080";
    window.location.href = `${baseUrl}/auth/google`;
  };

  const handleCompleteRegistration = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!chosenName.trim() || isInvalidLength) return;

    setError(null);
    try {
      await completeRegistration(chosenName.trim());
    } catch (err: any) {
      setError(err.message || 'Registration completion failed.');
    }
  };

  if (user?.pendingCompletion) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4">
        <div className="glass p-10 rounded-3xl w-full max-w-md card-shadow border border-white/5 animate-in fade-in zoom-in duration-300">
          <div className="text-center mb-10">
            <div className="inline-block p-3 rounded-2xl bg-secondary/10 mb-4">
              <UserPlus size={32} className="text-secondary" />
            </div>
            <h1 className="text-3xl font-black tracking-tight text-white mb-2">PICK YOUR ALIAS</h1>
            <p className="text-gray-400 text-sm">Every terrible person needs a name.</p>
          </div>

          <form onSubmit={handleCompleteRegistration} className="space-y-6">
            <div className="space-y-2">
              <div className="flex justify-between items-end px-1">
                <label className="text-xs font-bold text-gray-500 uppercase tracking-widest">Chosen Name</label>
                <span className={`text-[10px] font-bold ${isInvalidLength ? 'text-red-400' : 'text-gray-500'}`}>
                  {chosenName.length} / {maxLen}
                </span>
              </div>
              <input
                autoFocus
                type="text"
                placeholder="e.g. TotalAsshat"
                value={chosenName}
                onChange={(e) => setChosenName(e.target.value)}
                className={`w-full bg-white/5 border rounded-2xl px-6 py-4 text-white focus:outline-none focus:ring-2 transition-all font-bold text-lg ${
                  isInvalidLength ? 'border-red-500/50 focus:ring-red-500/30' : 'border-white/10 focus:ring-secondary/50'
                }`}
              />
              {isInvalidLength && (
                <p className="text-[10px] text-red-400 px-1 font-bold">
                  Name must be between {minLen} and {maxLen} characters.
                </p>
              )}
              <p className="text-[10px] text-gray-500 px-1 italic">
                A random 8-digit identifier will be appended to your name.
              </p>
            </div>

            {error && (
              <div className="p-4 bg-red-500/10 border border-red-500/20 rounded-2xl text-red-400 text-sm text-center">
                {error}
              </div>
            )}

            <div className="pt-2 flex gap-3">
              <button
                type="submit"
                disabled={!chosenName.trim() || isInvalidLength}
                className="w-full bg-secondary hover:bg-secondary/80 text-background py-4 rounded-2xl font-black transition-all disabled:opacity-50 disabled:cursor-not-allowed shadow-[0_0_20px_rgba(234,179,8,0.3)]"
              >
                COMPLETE REGISTRATION
              </button>
            </div>
          </form>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="glass p-10 rounded-3xl w-full max-w-md card-shadow border border-white/5">
        <div className="text-center mb-10">
          <div className="inline-block p-3 rounded-2xl bg-primary/10 mb-4">
            <LogIn size={32} className="text-primary" />
          </div>
          <h1 className="text-5xl font-black tracking-tighter text-white mb-2">RUTHLESS</h1>
          <p className="text-gray-400 uppercase text-xs tracking-[0.2em] font-bold">A Game for Terrible People</p>
        </div>

        {error && (
          <div className="mb-6 p-4 bg-red-500/10 border border-red-500/20 rounded-2xl text-red-400 text-sm text-center">
            {error}
          </div>
        )}
        
        <div className="space-y-8">
          <div className="space-y-4">
            <p className="text-sm text-center text-gray-400 leading-relaxed px-4">
              Sign in or register with your Google account to start your journey of terror.
            </p>
            <button 
              onClick={handleGoogleLogin}
              className="w-full bg-white text-background hover:bg-white/90 py-4 rounded-2xl font-black transition-all flex items-center justify-center gap-3 shadow-xl"
            >
              <svg className="w-5 h-5" viewBox="0 0 24 24">
                <path
                  fill="currentColor"
                  d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                />
                <path
                  fill="currentColor"
                  d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                />
                <path
                  fill="currentColor"
                  d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l3.66-2.84z"
                />
                <path
                  fill="currentColor"
                  d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
                />
                <path fill="none" d="M0 0h24v24H0z" />
              </svg>
              CONTINUE WITH GOOGLE
            </button>
          </div>
          <div className="h-px bg-gradient-to-r from-transparent via-white/10 to-transparent" />

          {isDevelopment && (
            <div className="space-y-3">
              <h2 className="text-xl font-bold text-white flex items-center gap-2">
                <span className="p-1 rounded bg-gray-500/10"><UserPlus size={16} className="text-gray-400" /></span>
                Developer Access
              </h2>
              <p className="text-sm text-gray-400 leading-relaxed">
                Backend is in development mode. Use a developer token to login.
              </p>
              <form onSubmit={async (e) => {
                e.preventDefault();
                const target = e.target as typeof e.target & {
                  token: { value: string };
                };
                const tokenVal = target.token.value.trim();
                if (tokenVal) {
                  setError(null);
                  try {
                    loginWithToken(tokenVal);
                  } catch (err: any) {
                    setError(err.message || 'Developer authentication failed.');
                  }
                }
              }} className="flex gap-2">
              <input
                name="token"
                type="text"
                placeholder="Name or Token"
                className="flex-1 bg-white/5 border border-white/10 rounded-xl px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all"
              />
              <button
                type="submit"
                className="bg-white/10 hover:bg-white/20 text-white px-4 py-2 rounded-xl text-sm font-bold transition-all"
              >
                GO
              </button>
            </form>
          </div>
        )}
      </div>

      <p className="mt-10 text-[10px] text-gray-600 text-center uppercase tracking-widest font-bold font-mono">
        Identity: {import.meta.env.VITE_GOOGLE_CLIENT_ID ? 'Configured' : 'Missing'} | Port: {window.location.port}
      </p>
    </div>
  </div>
);
};
