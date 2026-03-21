import { useState, useEffect } from 'react';
import { useAuth } from './context/AuthContext';
import { notificationClient, createOptions } from './api/client';
import { Login } from './components/Login';
import { Lobby } from './components/Lobby';
import { GameBoard } from './components/GameBoard';
import { DeckSubscribe } from './components/DeckSubscribe';
import { Friends } from './components/Friends';
import { LayoutDashboard, Library as LibraryIcon, Layers, Users as UsersIcon } from 'lucide-react';
import { Library as LibraryView } from './components/Library';

type ViewState = 'sessions' | 'library' | 'game' | 'deck_subscribe' | 'friends';

function App() {
  const { user, token, loading } = useAuth();
  
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
    if (path.startsWith('/library/')) {
      const tab = path.split('/')[2];
      return { view: 'library' as ViewState, sessionId: null, deckId: null, activeTab: tab || 'decks' };
    }
    if (path === '/library') return { view: 'library' as ViewState, sessionId: null, deckId: null, activeTab: 'decks' };
    if (path === '/decks') return { view: 'library' as ViewState, sessionId: null, deckId: null, activeTab: 'decks' };
    if (path === '/cards') return { view: 'library' as ViewState, sessionId: null, deckId: null, activeTab: 'cards' };
    if (path === '/friends') return { view: 'friends' as ViewState, sessionId: null, deckId: null, activeTab: null };
    return { view: 'sessions' as ViewState, sessionId: null, deckId: null, activeTab: null };
  };

  const initial = getInitialState();

  const [activeSessionId, setActiveSessionId] = useState<string | null>(initial.sessionId);
  const [activeDeckId, setActiveDeckId] = useState<string | null>(initial.deckId);
  const [activeTab, setActiveTab] = useState<string | null>(initial.activeTab);
  const [view, setViewState] = useState<ViewState>(initial.view);
  const [hasNotifications, setHasNotifications] = useState(false);

  const checkNotifications = async () => {
    if (!user || user.pendingCompletion) return;
    try {
      const res = await notificationClient.getNotifications({}, createOptions(token));
      const notifications = res.response.notifications || [];
      const count = notifications.reduce((acc: any, n: any) => acc + n.count, 0);
      const changed = (count > 0) !== hasNotifications;
      setHasNotifications(count > 0);
      if (changed && count > 0) {
        window.dispatchEvent(new Event('notifications-updated'));
      }
    } catch (err) {
      console.error('Failed to get notifications:', err);
    }
  };

  useEffect(() => {
    checkNotifications();
    const interval = setInterval(checkNotifications, 10000);
    const handleReset = () => setHasNotifications(false);
    window.addEventListener('notifications-reset', handleReset);
    return () => {
      clearInterval(interval);
      window.removeEventListener('notifications-reset', handleReset);
    };
  }, [user, token]);

  const setView = (newView: ViewState, sessionId: string | null = null, deckId: string | null = null, tab?: string) => {
    setViewState(newView);
    
    let newPath = '';
    if (newView === 'game') {
      newPath = `/game/${sessionId}`;
    } else if (newView === 'deck_subscribe') {
      newPath = `/decks/${deckId}/subscribe`;
    } else if (newView === 'library' && !deckId) {
      newPath = `/library/${tab || 'decks'}`;
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
    if (id) {
       setView('library', null, id, tab);
    } else {
       setView('library', null, null, 'decks');
    }
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
              setView('library', null, null, 'decks');
            }}
            className={`flex items-center gap-2 px-4 py-2 rounded-xl font-bold text-sm transition-all ${
              view === 'library' ? 'bg-primary text-background' : 'text-gray-400 hover:text-white'
            }`}
          >
            <LibraryIcon size={16} />
            LIBRARY
          </button>
          <button
            onClick={() => setView('friends')}
            className={`relative flex items-center gap-2 px-4 py-2 rounded-xl font-bold text-sm transition-all ${
              view === 'friends' ? 'bg-primary text-background' : 'text-gray-400 hover:text-white'
            }`}
          >
            <UsersIcon size={16} />
            FRIENDS
            {hasNotifications && (
              <span className="w-2 h-2 bg-red-500 rounded-full inline-block" />
            )}
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
        ) : view === 'library' ? (
          <LibraryView 
            initialTab={activeTab as any || 'decks'} 
            activeDeckId={activeDeckId} 
            activeTab={activeTab} 
            onSelectDeck={handleSelectDeck} 
          />
        ) : view === 'deck_subscribe' && activeDeckId ? (
          <DeckSubscribe deckId={activeDeckId} />
        ) : view === 'friends' ? (
          <Friends />
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
