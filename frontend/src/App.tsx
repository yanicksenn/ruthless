import { useState, useEffect } from 'react';
import { useAuth } from './context/AuthContext';
import { Login } from './components/Login';
import { Lobby } from './components/Lobby';
import { GameBoard } from './components/GameBoard';
import { Decks } from './components/Decks';
import { Cards } from './components/Cards';
import { DeckSubscribe } from './components/DeckSubscribe';
import { LayoutDashboard, Library, Layers } from 'lucide-react';

type ViewState = 'sessions' | 'decks' | 'cards' | 'game' | 'deck_subscribe';

function App() {
  const { user, loading } = useAuth();
  
  const getInitialState = () => {
    const path = window.location.pathname;
    if (path.startsWith('/game/')) {
      const id = path.split('/')[2];
      return { view: 'game' as ViewState, sessionId: id, deckId: null, activeTab: null };
    }
    if (path.startsWith('/decks/')) {
      const parts = path.split('/');
      if (parts[3] === 'subscribe') {
        return { view: 'deck_subscribe' as ViewState, sessionId: null, deckId: parts[2], activeTab: null };
      }
      if (parts[2]) {
        return { view: 'decks' as ViewState, sessionId: null, deckId: parts[2], activeTab: parts[3] || 'cards' };
      }
    }
    if (path === '/decks') return { view: 'decks' as ViewState, sessionId: null, deckId: null, activeTab: null };
    if (path === '/cards') return { view: 'cards' as ViewState, sessionId: null, deckId: null, activeTab: null };
    return { view: 'sessions' as ViewState, sessionId: null, deckId: null, activeTab: null };
  };

  const initial = getInitialState();

  const [activeSessionId, setActiveSessionId] = useState<string | null>(initial.sessionId);
  const [activeDeckId, setActiveDeckId] = useState<string | null>(initial.deckId);
  const [activeTab, setActiveTab] = useState<string | null>(initial.activeTab);
  const [view, setViewState] = useState<ViewState>(initial.view);

  const setView = (newView: ViewState, sessionId: string | null = null, deckId: string | null = null, tab?: string) => {
    setViewState(newView);
    
    let newPath = '';
    if (newView === 'game') {
      newPath = `/game/${sessionId}`;
    } else if (newView === 'deck_subscribe') {
      newPath = `/decks/${deckId}/subscribe`;
    } else if (newView === 'decks' && deckId) {
      newPath = `/decks/${deckId}${tab ? `/${tab}` : ''}`;
    } else {
      newPath = `/${newView === 'sessions' ? '' : newView}`;
    }

    if (window.location.pathname !== newPath) {
      window.history.pushState(null, '', newPath);
    }
  };

  useEffect(() => {
    const handlePopState = () => {
      const state = getInitialState();
      setViewState(state.view);
      setActiveSessionId(state.sessionId);
      setActiveDeckId(state.deckId);
      setActiveTab(state.activeTab);
    };
    window.addEventListener('popstate', handlePopState);
    return () => window.removeEventListener('popstate', handlePopState);
  }, []);

  useEffect(() => {
    if (!loading) {
      if (!user) {
        if (window.location.pathname !== '/login') {
          window.history.replaceState(null, '', '/login');
        }
      } else if (user.pendingCompletion) {
        if (window.location.pathname !== '/register') {
          window.history.replaceState(null, '', '/register');
        }
      } else {
        if (window.location.pathname === '/login' || window.location.pathname === '/register') {
          window.history.replaceState(null, '', '/');
          setViewState('sessions');
        }
      }
    }
  }, [user, loading]);

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
    setView('game', id);
  };

  const handleLeaveSession = () => {
    setActiveSessionId(null);
    setView('sessions');
  };

  const handleSelectDeck = (id: string | null) => {
    setActiveDeckId(id);
    const tab = id ? 'cards' : undefined;
    setActiveTab(tab || null);
    setView('decks', null, id, tab);
  };

  return (
    <main className="min-h-screen">
      {view !== 'game' && view !== 'deck_subscribe' && (
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
            onClick={() => {
              setActiveDeckId(null);
              setView('decks');
            }}
            className={`flex items-center gap-2 px-4 py-2 rounded-xl font-bold text-sm transition-all ${
              view === 'decks' && !activeDeckId ? 'bg-primary text-background' : 'text-gray-400 hover:text-white'
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
      )}

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
          <Decks activeDeckId={activeDeckId} activeTab={activeTab} onSelectDeck={handleSelectDeck} />
        ) : view === 'cards' ? (
          <Cards />
        ) : view === 'deck_subscribe' && activeDeckId ? (
          <DeckSubscribe deckId={activeDeckId} />
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
