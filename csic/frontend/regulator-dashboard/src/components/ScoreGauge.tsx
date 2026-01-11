interface ScoreGaugeProps {
  score: number
  size?: 'small' | 'medium' | 'large'
}

export default function ScoreGauge({ score, size = 'medium' }: ScoreGaugeProps) {
  const sizeClasses = {
    small: 'w-20 h-20',
    medium: 'w-28 h-28',
    large: 'w-36 h-36',
  }
  
  const fontSizes = {
    small: 'text-xl',
    medium: 'text-3xl',
    large: 'text-4xl',
  }
  
  const getTierColor = (score: number) => {
    if (score >= 90) return '#eab308' // Gold
    if (score >= 75) return '#94a3b8' // Silver
    if (score >= 60) return '#b45309' // Bronze
    if (score >= 40) return '#f59e0b' // At Risk
    return '#ef4444' // Critical
  }
  
  const getTierLabel = (score: number) => {
    if (score >= 90) return 'GOLD'
    if (score >= 75) return 'SILVER'
    if (score >= 60) return 'BRONZE'
    if (score >= 40) return 'AT RISK'
    return 'CRITICAL'
  }
  
  const circumference = 2 * Math.PI * 45
  const strokeDashoffset = circumference - (score / 100) * circumference
  
  return (
    <div className={`${sizeClasses[size]} relative`}>
      <svg className="w-full h-full transform -rotate-90">
        {/* Background circle */}
        <circle
          cx="50%"
          cy="50%"
          r="45%"
          fill="none"
          stroke="#334155"
          strokeWidth="8"
        />
        {/* Progress circle */}
        <circle
          cx="50%"
          cy="50%"
          r="45%"
          fill="none"
          stroke={getTierColor(score)}
          strokeWidth="8"
          strokeDasharray={circumference}
          strokeDashoffset={strokeDashoffset}
          strokeLinecap="round"
          className="transition-all duration-500 ease-out"
        />
      </svg>
      <div className="absolute inset-0 flex flex-col items-center justify-center">
        <span className={`${fontSizes[size]} font-bold text-white`}>
          {Math.round(score)}
        </span>
        <span className="text-xs text-slate-400 uppercase tracking-wider">
          {getTierLabel(score)}
        </span>
      </div>
    </div>
  )
}
