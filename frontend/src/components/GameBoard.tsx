import React, { useState, useEffect } from 'react';
import { gameClient, sessionClient, deckClient, createOptions } from '../api/client';
import { Game, GameState, Card, Session } from '../api/ruthless';
import { useAuth } from '../context/AuthContext';
import { ArrowLeft, Play, Crown, Check, Info, Users, Layers, LogOut } from 'lucide-react';
import { motion } from 'framer-motion';

interface GameBoardProps {
  sessionId: string;
  onBack: () => void;
  onLeave: () => void;
}

export const GameBoard: React.FC<GameBoardProps> = ({ sessionId, onBack, onLeave }) => {
  const { token, user } = useAuth();

  const [session, setSession] = useState<Session | null>(null);
  const [game, setGame] = useState<Game | null>(null);
  const [hand, setHand] = useState<Card[]>([]);
  const [selectedCards, setSelectedCards] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [decks, setDecks] = useState<{id: string, name: string}[]>([]);
  const [hoveredPlayId, setHoveredPlayId] = useState<string | null>(null);

  const isOwner = session?.ownerId === user?.id;

  const fetchData = async () => {
    try {
      const sResponse = await sessionClient.getSession({ id: sessionId }, createOptions(token));
      setSession(sResponse.response);

      try {
        const gResponse = await gameClient.getGameBySession({ sessionId }, createOptions(token));
        setGame(gResponse.response);
        
        if (gResponse.response.state === GameState.PLAYING || gResponse.response.state === GameState.JUDGING) {
           const hResponse = await gameClient.getHand({ gameId: gResponse.response.id }, createOptions(token));
           setHand(hResponse.response.cards || []);
        }
      } catch (gErr) {
        // Game might not be created yet, that's fine
        setGame(null);
      }

      // Fetch decks if owner and game is in WAITING state (or not yet fetched/null)
      // The game is now created automatically with the session, so it will be in WAITING state.
      if (isOwner && (game === null || game.state === GameState.WAITING)) {
        const dResponse = await deckClient.listDecks({}, createOptions(token));
        setDecks(dResponse.response.decks.map(d => ({ id: d.id, name: d.name })));
      }
    } catch (err) {
      console.error('Fetch error:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 3000);
    return () => clearInterval(interval);
  }, [sessionId, token, user?.id]);

  /* handleCreateGame removed - game is created automatically with session */


  const handleStartGame = async () => {
    if (!game) return;
    try {
      await gameClient.startGame({ id: game.id }, createOptions(token));
      fetchData();
    } catch (err) { console.error(err); }
  };

  const handleAddDeck = async (deckId: string) => {
    try {
       await sessionClient.addDeckToSession({ sessionId, deckId }, createOptions(token));
       fetchData();
    } catch (err) { console.error(err); }
  };

  const handlePlayCards = async () => {
    if (!game) return;
    try {
      await gameClient.playCards({ gameId: game.id, cardIds: selectedCards }, createOptions(token));
      setSelectedCards([]);
      fetchData();
    } catch (err) { console.error(err); }
  };

  const handleSelectWinner = async (playId: string) => {
    if (!game) return;
    try {
      await gameClient.selectWinner({ gameId: game.id, playId }, createOptions(token));
      fetchData();
    } catch (err) { console.error(err); }
  };

  if (loading && !session) return <div className="flex items-center justify-center min-h-screen">Loading...</div>;

  const currentRound = game?.rounds[game.rounds.length - 1];
  const isCzar = currentRound?.czarId === user?.id;

  const handleLeave = async () => {
    try {
      await sessionClient.leaveSession({ sessionId }, createOptions(token));
      onLeave();
    } catch (err) {
      console.error('Failed to leave session:', err);
      onLeave(); // Leave anyway on frontend
    }
  };

  const renderBlackCardText = (text: string) => {
    if (!text) return "...";
    const parts = text.split('___');
    if (parts.length === 1) return text;

    // Czar see only blanks during PLAYING phase
    const showBlanks = isCzar && game?.state === GameState.PLAYING;
    
    // During judging, Czar can see preview of hovered play
    const previewCards = (isCzar && game?.state === GameState.JUDGING && hoveredPlayId) 
      ? Object.values(currentRound?.plays || {}).find(p => p.id === hoveredPlayId)?.cards 
      : null;

    return (
      <>
        {parts.map((part, i) => (
          <React.Fragment key={i}>
            {part}
            {i < parts.length - 1 && (
              <span className={`inline-block border-b-2 px-2 min-w-[80px] text-center transition-all ${
                (!showBlanks && (selectedCards[i] || previewCards?.[i])) 
                  ? 'text-primary border-primary italic mx-1' 
                  : 'text-gray-600 border-gray-600 translate-y-1'
              }`}>
                {!showBlanks && (previewCards?.[i]?.text || (selectedCards[i] ? hand.find(c => c.id === selectedCards[i])?.text : ""))}
              </span>
            )}
          </React.Fragment>
        ))}
      </>
    );
  };

  return (
    <div className="min-h-screen flex flex-col">
      {/* Top Bar */}
      <nav className="glass border-b border-white/5 px-6 py-4 flex justify-between items-center sticky top-0 z-50">
        <div className="flex items-center gap-4">
          <button 
            onClick={onBack} 
            className="flex items-center gap-2 text-gray-400 hover:text-white transition-colors"
            title="Return to lobby without leaving session"
          >
            <ArrowLeft size={18} /> <span className="font-bold text-xs uppercase tracking-widest">Back to Lobby</span>
          </button>
          
          <div className="h-4 w-px bg-white/10 mx-2" />
          
          <button 
            onClick={handleLeave} 
            className="flex items-center gap-2 text-red-500/60 hover:text-red-500 transition-colors"
            title="Leave session entirely"
          >
            <LogOut size={16} /> <span className="font-bold text-[10px] uppercase tracking-tighter">Leave Session</span>
          </button>
        </div>
        <div className="flex items-center gap-6">
           <div className="flex items-center gap-2">
             <Users size={16} className="text-primary" />
             <span className="font-bold text-sm tracking-tight">{session?.playerIds.length} Players</span>
           </div>
           {game && (
             <div className="bg-white/5 px-3 py-1 rounded-full text-[10px] font-black uppercase text-primary border border-primary/20">
               {GameState[game.state].replace('GAME_STATE_', '')}
             </div>
           )}
        </div>
      </nav>


      <main className="flex-1 p-6 md:p-12">
        {!game ? (
          <div className="flex flex-col items-center justify-center min-h-[60vh] gap-4">
             <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-primary"></div>
             <p className="text-gray-500 font-bold uppercase tracking-widest text-xs">Summoning the Session...</p>
          </div>
        ) : game.state === GameState.WAITING ? (
           <div className="max-w-2xl mx-auto mt-12 space-y-8">
              <div className="text-center">
                <h2 className="text-4xl font-black mb-2 tracking-tighter uppercase">Lobby</h2>
                <p className="text-gray-400 font-bold tracking-widest uppercase text-xs italic">Gather your friends for the chaos</p>
              </div>

              <div className="glass p-8 rounded-3xl space-y-8 border-primary/20 border-2">
                 <div className="flex justify-between items-start">
                    <div>
                        <h3 className="text-2xl font-black mb-1">READY TO START?</h3>
                        <p className="text-gray-400 text-sm font-bold uppercase tracking-wider">
                          {session?.playerIds.length || 0} Players Joined
                        </p>
                    </div>
                    <div className="bg-primary/20 p-3 rounded-full border border-primary/40 text-primary">
                       <Play size={24} fill="currentColor" />
                    </div>
                 </div>

                 {game.minRequiredPlayers > (session?.playerIds.length || 0) && (
                   <div className="flex items-center justify-center gap-2 text-red-500 font-bold bg-red-500/10 p-3 rounded-xl border border-red-500/20">
                     <Info size={16} />
                     <span className="text-xs uppercase tracking-tight">Need {game.minRequiredPlayers - (session?.playerIds.length || 0)} more players to start</span>
                   </div>
                 )}
                 
                 {isOwner && (
                    <div className="space-y-4 pt-4 border-t border-white/5">
                      <h3 className="flex items-center gap-2 font-black uppercase text-xs tracking-widest text-gray-500">
                        <Layers size={14} /> Active Decks ({session?.deckIds.length || 0})
                      </h3>
                      <div className="grid grid-cols-2 gap-2">
                         {decks.map(d => (
                           <button 
                             key={d.id} 
                             onClick={() => handleAddDeck(d.id)}
                             disabled={session?.deckIds.includes(d.id)}
                             className={`p-3 rounded-xl border text-xs font-bold transition-all text-left flex justify-between items-center ${
                               session?.deckIds.includes(d.id) 
                               ? 'bg-primary/20 border-primary/40 text-primary' 
                               : 'bg-white/5 border-white/10 hover:border-white/20'
                             }`}
                           >
                             <span className="truncate mr-2">{d.name}</span>
                             {session?.deckIds.includes(d.id) && <Check size={12} />}
                           </button>
                         ))}
                      </div>
                    </div>
                 )}

                 <div className="space-y-4 pt-4 border-t border-white/5">
                    <h3 className="flex items-center gap-2 font-black uppercase text-xs tracking-widest text-gray-500">
                      <Users size={14} /> Joined Players ({session?.playerIds.length || 0})
                    </h3>
                    <div className="grid grid-cols-2 gap-2">
                       {(game.players || []).map(p => (
                         <div key={p.id} className="p-3 rounded-xl bg-white/5 border border-white/10 text-xs font-bold flex items-center justify-between">
                            <span className="truncate">
                              {p.name}
                              {p.identifier && <span className="text-gray-500 italic">#{p.identifier}</span>}
                            </span>
                            {p.id === session?.ownerId && <Crown size={12} className="text-primary flex-shrink-0" />}
                         </div>
                       ))}
                    </div>
                 </div>

                 {isOwner ? (
                    <button 
                      onClick={handleStartGame} 
                      disabled={(session?.playerIds.length || 0) < game.minRequiredPlayers || !session?.deckIds.length}
                      className="w-full bg-primary hover:bg-primary-dark disabled:opacity-30 disabled:cursor-not-allowed text-white font-black py-4 rounded-2xl transition-all shadow-lg shadow-primary/20"
                    >
                      {!session?.deckIds.length ? "SELECT AT LEAST ONE DECK" : "START FIRST ROUND"}
                    </button>
                 ) : (
                    <div className="text-center py-4">
                        <p className="text-primary font-black uppercase tracking-widest text-sm animate-pulse">Waiting for host to start...</p>
                    </div>
                 )}
              </div>
           </div>
        ) : (
           <div className="flex flex-col lg:flex-row gap-8">
             <div className="flex-1 space-y-12">
                {/* Black Card Display */}
                <div className="flex justify-center">
                   <motion.div 
                     layoutId="black-card"
                     className="bg-black border border-white/20 p-8 rounded-2xl w-full max-w-sm aspect-[3/4] flex flex-col justify-between card-shadow shadow-primary/20"
                   >
                      <p className="text-2xl font-black leading-tight tracking-tight">
                         {renderBlackCardText(currentRound?.blackCard?.text || "")}
                      </p>
                      <div className="flex justify-between items-end">
                         <div className="text-[10px] font-black tracking-[0.2em] opacity-30">RUTHLESS</div>
                         <Crown className={isCzar ? "text-primary fill-primary" : "text-gray-800"} size={24} />
                      </div>
                   </motion.div>
                </div>

                {/* Status Indicator */}
                <div className="text-center space-y-2">
                   {isCzar ? (
                     <div className="inline-flex items-center gap-2 bg-primary/10 border border-primary/20 px-4 py-1.5 rounded-full text-primary font-black text-[10px] uppercase tracking-widest">
                        <Crown size={12} fill="currentColor" /> You are the Czar
                     </div>
                   ) : (
                     <div className="inline-flex items-center gap-2 bg-white/5 border border-white/10 px-4 py-1.5 rounded-full text-gray-400 font-black text-[10px] uppercase tracking-widest">
                        <Users size={12} /> Submit your play
                     </div>
                   )}
                   <h3 className="text-xl font-bold italic text-gray-500">
                      {game.state === GameState.PLAYING ? "Players are choosing..." : "Czar is judging..."}
                   </h3>
                </div>

                {/* Submissions or Hand */}
                <div className="mt-12">
                   {game.state === GameState.JUDGING ? (
                     <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                        {Object.entries(currentRound?.plays || {}).map(([pid, play]) => (
                          <div 
                             key={pid} 
                             className={`glass p-6 rounded-2xl flex flex-col justify-between aspect-[4/3] relative group transition-all ${isCzar ? 'hover:border-primary cursor-pointer' : ''}`}
                             onClick={() => isCzar && handleSelectWinner(play.id)}
                             onMouseEnter={() => isCzar && setHoveredPlayId(play.id)}
                             onMouseLeave={() => isCzar && setHoveredPlayId(null)}
                          >
                             <div className="space-y-2">
                                {play.cards.map(c => (
                                  <p key={c.id} className="text-lg font-bold">{c.text}</p>
                                ))}
                             </div>
                             {isCzar && (
                               <div className="absolute inset-0 bg-primary/10 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity rounded-2xl">
                                  <span className="bg-primary text-white font-black px-4 py-2 rounded-lg text-sm">ELECT WINNER</span>
                               </div>
                             )}
                          </div>
                        ))}
                     </div>
                   ) : !isCzar ? (
                     <div className="space-y-6">
                        <div className="flex justify-between items-center">
                           <h3 className="font-black text-xs uppercase tracking-[0.2em] text-gray-500">Your Hand</h3>
                           <button 
                             onClick={handlePlayCards}
                             disabled={selectedCards.length === 0}
                             className="bg-primary hover:bg-primary-dark disabled:opacity-30 px-6 py-2 rounded-xl text-xs font-black transition-all"
                           >
                             SUBMIT SELECTION
                           </button>
                        </div>
                        <div className="grid grid-cols-2 md:grid-cols-5 gap-3">
                           {(hand || []).map((card) => (
                             <div 
                                key={card.id}
                                onClick={() => {
                                  if (selectedCards.includes(card.id)) {
                                    setSelectedCards(s => s.filter(id => id !== card.id));
                                  } else {
                                    const blankCount = (currentRound?.blackCard?.text.split('___').length || 1) - 1;
                                    if (selectedCards.length < Math.max(1, blankCount)) {
                                      setSelectedCards(s => [...s, card.id]);
                                    }
                                  }
                                }}
                                className={`p-4 rounded-xl border aspect-[1/1] text-sm font-bold transition-all flex flex-col justify-between cursor-pointer ${
                                  selectedCards.includes(card.id) 
                                  ? 'bg-white text-black border-white scale-105 z-10' 
                                  : 'bg-surface border-white/5 text-gray-300 hover:border-white/20'
                                }`}
                             >
                                <p className="line-clamp-4">{card.text}</p>
                                <div className="flex justify-between items-center">
                                  <span className="text-[10px] font-black opacity-40">
                                    {selectedCards.indexOf(card.id) !== -1 ? `#${selectedCards.indexOf(card.id) + 1}` : ''}
                                  </span>
                                  <div className="opacity-20"><Info size={12} /></div>
                                </div>
                             </div>
                           ))}
                        </div>
                     </div>
                   ) : (
                      <div className="bg-white/5 border border-dashed border-white/10 p-12 rounded-3xl text-center">
                         <p className="text-gray-500 font-bold italic">Wait for the plebs to finish their selections...</p>
                      </div>
                   )}
                </div>
             </div>

             {/* Right side container for players */}
             <div className="w-full lg:w-72 space-y-4">
                <div className="glass p-6 rounded-3xl border-primary/20 border-2 space-y-4 sticky top-24">
                   <h3 className="flex items-center gap-2 font-black uppercase text-xs tracking-widest text-gray-500 mb-4">
                     <Users size={14} /> Players
                   </h3>
                   <div className="space-y-2">
                      {(game.players || []).map(p => {
                        const isCzarNow = p.id === currentRound?.czarId;
                        return (
                          <div key={p.id} className={`p-3 rounded-xl border flex justify-between items-center ${isCzarNow ? 'bg-primary/20 border-primary/40 text-primary' : 'bg-white/5 border-white/10 text-white'}`}>
                             <div className="flex items-center gap-2 truncate">
                               {isCzarNow && <Crown size={14} className="flex-shrink-0" />}
                               <span className="font-bold text-sm truncate">
                                 {p.name}
                                 {p.identifier && <span className="text-gray-500 italic">#{p.identifier}</span>}
                               </span>
                             </div>
                             <div className="font-black text-xs bg-black/40 px-2 py-1 rounded-md">
                                {game.scores[p.id] || 0}
                             </div>
                          </div>
                        );
                      })}
                   </div>
                </div>
             </div>
           </div>
        )}
      </main>
    </div>
  );
};
