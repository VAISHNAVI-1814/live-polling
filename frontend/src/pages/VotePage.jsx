import React, { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api } from '../services/api';
import confetti from 'canvas-confetti';
import { 
  Vote, 
  CheckCircle2, 
  AlertCircle, 
  Clock, 
  BarChart3, 
  ArrowRight,
  ShieldCheck
} from 'lucide-react';
import LiveResultBar from '../components/LiveResultBar';
import { usePollWebSocket } from '../hooks/usePollWebSocket';

export default function VotePage() {
  const { shareCode } = useParams();
  const [poll, setPoll] = useState(null);
  const [results, setResults] = useState(null);
  const [selectedOption, setSelectedOption] = useState('');
  const [hasVoted, setHasVoted] = useState(false);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [successMessage, setSuccessMessage] = useState('');

  // WebSocket hook for live updates once voted
  usePollWebSocket(poll?.id, (updatedResults) => {
    if (updatedResults) {
      setResults(updatedResults);
    }
  });

  useEffect(() => {
    async function loadPoll() {
      try {
        setLoading(true);
        setError('');
        const pollData = await api.getPollByShareCode(shareCode);
        setPoll(pollData);

        // Check if visitor has already cast a vote
        const votedCheck = await api.checkIfVoted(pollData.id);
        if (votedCheck.hasVoted) {
          setHasVoted(true);
          const resultsData = await api.getResults(pollData.id);
          setResults(resultsData);
        }
      } catch (err) {
        setError(err.message || 'Failed to load poll');
      } finally {
        setLoading(false);
      }
    }

    if (shareCode) {
      loadPoll();
    }
  }, [shareCode]);

  const handleVote = async (e) => {
    e.preventDefault();
    if (!selectedOption) {
      setError('Please select an option to vote');
      return;
    }

    try {
      setSubmitting(true);
      setError('');
      const res = await api.vote(poll.id, selectedOption);
      setSuccessMessage('Your vote has been recorded live!');
      setHasVoted(true);
      setResults(res.results);

      // Trigger confetti celebration
      try {
        confetti({
          particleCount: 80,
          spread: 70,
          origin: { y: 0.6 },
          colors: ['#6366f1', '#10b981', '#f59e0b', '#ec4899']
        });
      } catch (e) {
        // Confetti fallback
      }
    } catch (err) {
      if (err.status === 409) {
        setHasVoted(true);
        setError('You have already voted on this poll.');
        const freshResults = await api.getResults(poll.id);
        setResults(freshResults);
      } else {
        setError(err.message || 'Failed to record vote');
      }
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-[80vh] flex items-center justify-center">
        <div className="animate-spin rounded-full h-10 w-10 border-t-2 border-b-2 border-indigo-500"></div>
      </div>
    );
  }

  if (error && !poll) {
    return (
      <div className="max-w-md mx-auto px-4 py-20 text-center">
        <div className="w-12 h-12 rounded-2xl bg-rose-500/10 text-rose-400 flex items-center justify-center mx-auto mb-4 border border-rose-500/20">
          <AlertCircle className="w-6 h-6" />
        </div>
        <h2 className="text-xl font-bold text-white mb-2">Poll Not Found</h2>
        <p className="text-sm text-slate-400 mb-6">
          The poll you are looking for might have been removed or the share code is incorrect.
        </p>
        <Link
          to="/"
          className="inline-flex items-center gap-2 px-4 py-2 text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-500 rounded-xl transition-colors"
        >
          Go to Home
        </Link>
      </div>
    );
  }

  const isClosed = poll.status === 'closed';
  const isExpired = poll.expires_at && new Date(poll.expires_at) < new Date();

  return (
    <div className="max-w-2xl mx-auto px-4 sm:px-6 py-10">
      {/* Poll Card */}
      <div className="bg-slate-900/90 border border-slate-800 rounded-3xl p-6 sm:p-8 shadow-2xl backdrop-blur-sm relative overflow-hidden">
        {/* Glow Header Accent */}
        <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-indigo-500 via-purple-500 to-emerald-500" />

        {/* Top Badges */}
        <div className="flex flex-wrap items-center justify-between gap-2 mb-6">
          <div className="flex items-center gap-2">
            <span className="font-mono text-xs px-2.5 py-1 rounded-full bg-slate-950 text-indigo-400 border border-slate-800 font-semibold">
              #{poll.share_code}
            </span>
            <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold ${
              !isClosed && !isExpired
                ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                : 'bg-slate-800 text-slate-400 border border-slate-700'
            }`}>
              <span className={`w-1.5 h-1.5 rounded-full ${!isClosed && !isExpired ? 'bg-emerald-400 animate-pulse' : 'bg-slate-500'}`} />
              {isExpired ? 'Expired' : isClosed ? 'Closed' : 'Active'}
            </span>
          </div>

          {poll.expires_at && (
            <span className="flex items-center gap-1 text-xs text-slate-400">
              <Clock className="w-3.5 h-3.5 text-slate-500" />
              {isExpired ? 'Voting Closed' : `Closes ${new Date(poll.expires_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`}
            </span>
          )}
        </div>

        {/* Question Title */}
        <h1 className="text-xl sm:text-2xl font-bold text-white mb-6 leading-tight">
          {poll.question}
        </h1>

        {/* Error or Notice */}
        {error && (
          <div className="mb-6 p-4 rounded-xl bg-rose-500/10 border border-rose-500/20 text-rose-400 text-sm flex items-center gap-3">
            <AlertCircle className="w-5 h-5 flex-shrink-0" />
            <span>{error}</span>
          </div>
        )}

        {/* Success Notice */}
        {successMessage && (
          <div className="mb-6 p-4 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-sm flex items-center gap-3 animate-fadeIn">
            <CheckCircle2 className="w-5 h-5 flex-shrink-0" />
            <span>{successMessage}</span>
          </div>
        )}

        {/* Closed Banner */}
        {(isClosed || isExpired) && !hasVoted && (
          <div className="mb-6 p-4 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-300 text-sm flex items-center justify-between gap-3">
            <div className="flex items-center gap-2.5">
              <Clock className="w-4 h-4 text-amber-400 flex-shrink-0" />
              <span>This poll is closed and no longer accepting votes.</span>
            </div>
            <Link
              to={`/poll/${poll.id}/results`}
              className="px-3 py-1.5 rounded-lg bg-amber-500/20 text-amber-200 text-xs font-semibold hover:bg-amber-500/30 transition-colors flex-shrink-0"
            >
              View Results
            </Link>
          </div>
        )}

        {/* Voting Form (if active & not voted) */}
        {!hasVoted && !isClosed && !isExpired ? (
          <form onSubmit={handleVote} className="space-y-4">
            <div className="space-y-2.5">
              {poll.options.map((opt) => {
                const isSelected = selectedOption === opt.id;
                return (
                  <label
                    key={opt.id}
                    className={`flex items-center justify-between p-4 rounded-2xl border cursor-pointer transition-all duration-200 ${
                      isSelected
                        ? 'bg-indigo-600/15 border-indigo-500 text-white shadow-md shadow-indigo-600/10'
                        : 'bg-slate-950/60 border-slate-800/80 hover:border-slate-700 text-slate-200 hover:bg-slate-950'
                    }`}
                  >
                    <div className="flex items-center gap-3.5">
                      <div className={`w-5 h-5 rounded-full border flex items-center justify-center transition-colors ${
                        isSelected ? 'border-indigo-500 bg-indigo-600' : 'border-slate-600 bg-transparent'
                      }`}>
                        {isSelected && <div className="w-2 h-2 rounded-full bg-white" />}
                      </div>
                      <span className="font-medium text-sm sm:text-base">{opt.text}</span>
                    </div>
                    <input
                      type="radio"
                      name="poll-option"
                      value={opt.id}
                      checked={isSelected}
                      onChange={() => setSelectedOption(opt.id)}
                      className="sr-only"
                    />
                  </label>
                );
              })}
            </div>

            <button
              type="submit"
              disabled={submitting || !selectedOption}
              className="w-full mt-6 py-3 px-6 rounded-2xl bg-indigo-600 hover:bg-indigo-500 active:scale-[0.99] text-white font-semibold text-sm shadow-lg shadow-indigo-600/30 transition-all flex items-center justify-center gap-2 disabled:opacity-40 disabled:cursor-not-allowed"
            >
              {submitting ? (
                <div className="animate-spin rounded-full h-5 w-5 border-2 border-white border-t-transparent" />
              ) : (
                <>
                  Cast Live Vote
                  <ArrowRight className="w-4 h-4" />
                </>
              )}
            </button>
          </form>
        ) : (
          /* Live Results View */
          <div className="space-y-6 pt-2">
            <div className="flex items-center justify-between text-xs text-slate-400 pb-2 border-b border-slate-800">
              <span className="flex items-center gap-1.5 text-emerald-400 font-medium">
                <ShieldCheck className="w-4 h-4" />
                {hasVoted ? 'You have voted' : 'Poll Closed'}
              </span>
              <span className="font-mono">
                {results?.total_votes || 0} Total Live Votes
              </span>
            </div>

            <div className="space-y-3">
              {(results?.options || poll.options).map((opt) => {
                const maxVotes = Math.max(...(results?.options || []).map(o => o.votes || 0), 0);
                const isLeader = maxVotes > 0 && opt.votes === maxVotes;
                return (
                  <LiveResultBar
                    key={opt.id || opt.option_id}
                    option={opt}
                    isLeader={isLeader}
                    totalVotes={results?.total_votes || 0}
                  />
                );
              })}
            </div>

            <div className="pt-4 flex items-center justify-between">
              <Link
                to={`/poll/${poll.id}/results`}
                className="inline-flex items-center gap-2 text-xs font-semibold text-indigo-400 hover:text-indigo-300 transition-colors"
              >
                <BarChart3 className="w-4 h-4" />
                View Full Live Results Page
              </Link>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
