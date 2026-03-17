import { useState } from 'react';
import { useAuth } from './context/AuthContext';
import { Login } from './components/Login';
import { Lobby } from './components/Lobby';
import { GameBoard } from './components/GameBoard';
import { Decks } from './components/Decks';
import { Cards } from './components/Cards';
import { LayoutDashboard, Library, Layers } from 'lucide-react';


function App() {
  const { user, loading } = useAuth();
  const [activeSessionId, setActiveSessionId] = useState<string | null>(null);
  const [view, setView] = useState<'sessions' | 'decks' | 'cards' | 'game'>('sessions');

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-primary"></div>
      </div>
    );
  }

  if (!user || user.pendingCompletion) {
    return <Login />;
  }

  const handleJoinSession = (id: string) => {
    setActiveSessionId(id);
    setView('game');
  };

  const handleLeaveSession = () => {
    setActiveSessionId(null);
    setView('sessions');
  };

  return (
    <main className="min-h-screen">
      {/* Global Nav */}
      <nav className="fixed bottom-8 left-1/2 -translate-x-1/2 glass px-4 py-2 rounded-2xl border border-white/10 flex gap-2 z-50">
        <button
          onClick={() => setView('sessions')}
          className={`flex items-center gap-2 px-4 py-2 rounded-xl font-bold text-sm transition-all ${
            view === 'sessions' ? 'bg-primary text-background' : 'text-gray-400 hover:text-white'
          }`}
        >
          <LayoutDashboard size={16} />
          SESSIONS
        </button>

        <button
          onClick={() => setView('decks')}
          className={`flex items-center gap-2 px-4 py-2 rounded-xl font-bold text-sm transition-all ${
            view === 'decks' ? 'bg-primary text-background' : 'text-gray-400 hover:text-white'
          }`}
        >
          <Library size={16} />
          DECKS
        </button>
        <button
          onClick={() => setView('cards')}
          className={`flex items-center gap-2 px-4 py-2 rounded-xl font-bold text-sm transition-all ${
            view === 'cards' ? 'bg-primary text-background' : 'text-gray-400 hover:text-white'
          }`}
        >
          <Layers size={16} />
          CARDS
        </button>
      </nav>

      <div className="pb-24">
        {view === 'sessions' ? (
          <Lobby 
            onJoinSession={handleJoinSession} 
            activeSessionId={activeSessionId}
          />
        ) : view === 'game' && activeSessionId ? (
          <GameBoard 
            sessionId={activeSessionId} 
            onBack={() => setView('sessions')}
            onLeave={handleLeaveSession}
          />
        ) : view === 'decks' ? (
          <Decks />
        ) : view === 'cards' ? (
          <Cards />
        ) : (
          /* Fallback if somehow in game view without active session */
          <Lobby 
            onJoinSession={handleJoinSession} 
            activeSessionId={activeSessionId}
          />
        )}
      </div>
    </main>
  );
}


export default App;
