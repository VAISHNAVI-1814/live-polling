import React from 'react';
import { Award } from 'lucide-react';

export default function LiveResultBar({ option, isLeader, totalVotes }) {
  const pct = Math.round((option.percentage || 0) * 10) / 10;

  return (
    <div className="group relative bg-slate-900/90 border border-slate-800 hover:border-slate-700 rounded-xl p-4 transition-all duration-200">
      {/* Background Animated Progress Fill */}
      <div
        className={`absolute top-0 bottom-0 left-0 rounded-xl transition-all duration-700 ease-out opacity-20 ${
          isLeader && totalVotes > 0
            ? 'bg-gradient-to-r from-emerald-500 to-teal-400'
            : 'bg-gradient-to-r from-indigo-500 to-purple-500'
        }`}
        style={{ width: `${Math.max(pct, 0)}%` }}
      />

      <div className="relative z-10 flex items-center justify-between gap-4">
        <div className="flex items-center gap-2.5 min-w-0">
          {isLeader && totalVotes > 0 && (
            <span className="p-1 rounded-md bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 flex-shrink-0" title="Current Leader">
              <Award className="w-4 h-4" />
            </span>
          )}
          <span className="font-medium text-slate-200 truncate group-hover:text-white">
            {option.text}
          </span>
        </div>

        <div className="flex items-center gap-3 flex-shrink-0">
          <span className="text-xs text-slate-400 font-mono">
            {option.votes} {option.votes === 1 ? 'vote' : 'votes'}
          </span>
          <span className={`text-sm font-bold font-mono px-2 py-0.5 rounded-md ${
            isLeader && totalVotes > 0
              ? 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/30'
              : 'bg-indigo-500/20 text-indigo-300 border border-indigo-500/30'
          }`}>
            {pct}%
          </span>
        </div>
      </div>

      {/* Thin bottom bar indicator */}
      <div className="w-full bg-slate-800/80 rounded-full h-1.5 mt-3 overflow-hidden">
        <div
          className={`h-full rounded-full transition-all duration-700 ease-out ${
            isLeader && totalVotes > 0
              ? 'bg-gradient-to-r from-emerald-500 to-teal-400 shadow-sm shadow-emerald-500/50'
              : 'bg-gradient-to-r from-indigo-500 to-purple-500 shadow-sm shadow-indigo-500/50'
          }`}
          style={{ width: `${Math.max(pct, 0)}%` }}
        />
      </div>
    </div>
  );
}
