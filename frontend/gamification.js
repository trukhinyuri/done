// Ninstyle gamification system - Database backed version
"use strict";

(function(window) {
    
    // Points system - based on task complexity and time spent
    const POINTS = {
        taskComplete: 100,         // Base points for completing a task
        onTimeBonus: 50,          // Bonus for completing before deadline
        streakBonus: 25,          // Bonus per day of streak (max 10 days)
        timeBonus: 10,            // Bonus per hour spent on task
        efficiencyBonus: 50,      // Bonus for completing faster than estimated
        levelUp: 500              // Bonus for leveling up
    };
    
    // Level thresholds
    const LEVELS = [
        { level: 1, points: 0, title: "Done" },
        { level: 2, points: 500, title: "Apprentice" },
        { level: 3, points: 1500, title: "Journeyman" },
        { level: 4, points: 3000, title: "Expert" },
        { level: 5, points: 5000, title: "Master" },
        { level: 6, points: 8000, title: "Champion" },
        { level: 7, points: 12000, title: "Hero" },
        { level: 8, points: 17000, title: "Legend" },
        { level: 9, points: 25000, title: "Mythic" },
        { level: 10, points: 50000, title: "Deity" }
    ];
    
    // Achievements
    const ACHIEVEMENTS = {
        firstTask: { id: 'firstTask', name: 'First Steps', description: 'Complete your first task', icon: '🎯' },
        streak3: { id: 'streak3', name: 'On Fire', description: '3 day streak', icon: '🔥' },
        streak7: { id: 'streak7', name: 'Week Warrior', description: '7 day streak', icon: '⚡' },
        streak30: { id: 'streak30', name: 'Monthly Master', description: '30 day streak', icon: '🌟' },
        points1000: { id: 'points1000', name: 'Point Collector', description: 'Earn 1000 points', icon: '💎' },
        points5000: { id: 'points5000', name: 'Point Master', description: 'Earn 5000 points', icon: '👑' },
        speedDemon: { id: 'speedDemon', name: 'Speed Demon', description: 'Complete 5 tasks in one day', icon: '⚡' },
        earlyBird: { id: 'earlyBird', name: 'Early Bird', description: 'Complete a task before deadline', icon: '🌅' }
    };
    
    // Cached gamification data
    let cachedData = null;
    let dataFetchPromise = null;
    
    // Fetch gamification data from API
    async function fetchGamificationData() {
        if (dataFetchPromise) {
            return dataFetchPromise;
        }
        
        dataFetchPromise = fetch('/api/getGamification')
            .then(response => {
                if (!response.ok) {
                    throw new Error('Failed to fetch gamification data');
                }
                return response.json();
            })
            .then(data => {
                cachedData = data;
                dataFetchPromise = null;
                return data;
            })
            .catch(error => {
                console.error('Error fetching gamification data:', error);
                dataFetchPromise = null;
                // Return default data if fetch fails
                return {
                    total_points: 0,
                    current_streak: 0,
                    longest_streak: 0,
                    level: 1,
                    completed_tasks: 0,
                    achievements: []
                };
            });
            
        return dataFetchPromise;
    }
    
    // Get current level based on points
    function getCurrentLevel(points) {
        let currentLevel = LEVELS[0];
        for (let i = LEVELS.length - 1; i >= 0; i--) {
            if (points >= LEVELS[i].points) {
                currentLevel = LEVELS[i];
                break;
            }
        }
        return currentLevel;
    }
    
    // Get progress to next level
    function getLevelProgress(points, completedTasks, firstTaskDate) {
        const currentLevel = getCurrentLevel(points);
        const nextLevelIndex = currentLevel.level; // This is already the index for next level
        const nextLevel = LEVELS[nextLevelIndex];
        
        if (!nextLevel || nextLevelIndex >= LEVELS.length) {
            return { percent: 100, pointsNeeded: 0, daysToNext: 0 };
        }
        
        const currentLevelPoints = currentLevel.points;
        const nextLevelPoints = nextLevel.points;
        const progress = points - currentLevelPoints;
        const needed = nextLevelPoints - currentLevelPoints;
        const percent = needed > 0 ? Math.floor((progress / needed) * 100) : 0;
        const pointsNeeded = Math.max(0, nextLevelPoints - points);
        
        // Calculate average points per day
        let avgPointsPerDay = 0;
        let daysToNext = null; // Use null to indicate no prediction available
        
        // Only calculate days prediction if user has completed tasks and has history
        if (completedTasks > 0 && firstTaskDate && points > 0) {
            const daysSinceFirst = Math.max(1, Math.floor((Date.now() - new Date(firstTaskDate).getTime()) / (1000 * 60 * 60 * 24)));
            avgPointsPerDay = points / daysSinceFirst;
            if (avgPointsPerDay > 0) {
                daysToNext = Math.ceil(pointsNeeded / avgPointsPerDay);
            }
        }
        // Don't provide estimate if no history - user needs to complete tasks first
        
        return {
            percent: Math.min(percent, 100),
            pointsNeeded: pointsNeeded,
            daysToNext: daysToNext // Will be null if no data available
        };
    }
    
    // Show achievement notification
    function showAchievement(achievement) {
        const notification = document.createElement('div');
        notification.className = 'achievement-notification';
        notification.innerHTML = `
            <div class="achievement-icon">${achievement.icon}</div>
            <div class="achievement-content">
                <div class="achievement-title">Achievement Unlocked!</div>
                <div class="achievement-name">${achievement.name}</div>
                <div class="achievement-description">${achievement.description}</div>
            </div>
        `;
        
        document.body.appendChild(notification);
        
        // Animate in
        setTimeout(() => {
            notification.classList.add('show');
            if (window.NinstyleSounds) {
                window.NinstyleSounds.taskComplete();
            }
        }, 100);
        
        // Remove after 4 seconds
        setTimeout(() => {
            notification.classList.remove('show');
            setTimeout(() => notification.remove(), 500);
        }, 4000);
    }
    
    // Update points display
    async function updatePointsDisplay() {
        const data = await fetchGamificationData();
        
        const totalPoints = data.total_points || 0;
        const completedTasks = data.completed_tasks || 0;
        const firstTaskDate = data.first_task_date;
        const currentLevel = getCurrentLevel(totalPoints);
        const progress = getLevelProgress(totalPoints, completedTasks, firstTaskDate);
        
        // Create or update points display
        let pointsDisplay = document.querySelector('.points-display');
        if (!pointsDisplay) {
            pointsDisplay = document.createElement('div');
            pointsDisplay.className = 'points-display';
            const header = document.querySelector('.page');
            if (header) {
                header.insertBefore(pointsDisplay, header.firstChild);
            }
        }
        
        const soundEnabled = window.NinstyleSounds && window.NinstyleSounds.isEnabled();
        
        pointsDisplay.innerHTML = `
            <div class="points-content">
                <div class="points-level">
                    <span class="level-badge">Lv.${currentLevel.level}</span>
                    <span class="level-title">${currentLevel.title}</span>
                </div>
                <div class="points-tasks">
                    <span class="tasks-icon">⭐</span>
                    <span class="tasks-value">${completedTasks}</span>
                </div>
                <div class="points-score">
                    <span class="points-icon">🏆</span>
                    <span class="points-value">${totalPoints.toLocaleString()}</span>
                </div>
                <div class="level-progress">
                    <div class="progress-bar">
                        <div class="progress-fill" style="width: ${progress.percent}%"></div>
                    </div>
                    <div class="progress-text">${progress.pointsNeeded} points to next level${progress.daysToNext !== null && progress.daysToNext > 0 ? ` (~${progress.daysToNext} day${progress.daysToNext !== 1 ? 's' : ''})` : ''}</div>
                </div>
                <button class="sound-toggle ${soundEnabled ? 'enabled' : ''}" title="Toggle sound effects">
                    <span class="sound-icon">${soundEnabled ? '🔊' : '🔇'}</span>
                </button>
            </div>
        `;
        
        // Add sound toggle event listener
        const soundToggle = pointsDisplay.querySelector('.sound-toggle');
        if (soundToggle) {
            soundToggle.addEventListener('click', function() {
                const enabled = window.NinstyleSounds.toggleSound();
                this.classList.toggle('enabled', enabled);
                this.querySelector('.sound-icon').textContent = enabled ? '🔊' : '🔇';
            });
        }
    }
    
    // Handle task completion - Points are now calculated on the backend
    async function onTaskComplete() {
        // Wait a bit for the backend to update
        setTimeout(async () => {
            // Clear cache to force refresh
            cachedData = null;
            
            // Fetch updated data
            const data = await fetchGamificationData();
            
            // Check achievements based on the new data
            const achievements = data.achievements || [];
            const totalPoints = data.total_points || 0;
            const currentStreak = data.current_streak || 0;
            const newAchievements = [];
            
            // Check for new achievements
            if (data.completed_tasks === 1 && !achievements.includes('firstTask')) {
                newAchievements.push(ACHIEVEMENTS.firstTask);
            }
            
            if (currentStreak >= 3 && !achievements.includes('streak3')) {
                newAchievements.push(ACHIEVEMENTS.streak3);
            }
            if (currentStreak >= 7 && !achievements.includes('streak7')) {
                newAchievements.push(ACHIEVEMENTS.streak7);
            }
            if (currentStreak >= 30 && !achievements.includes('streak30')) {
                newAchievements.push(ACHIEVEMENTS.streak30);
            }
            
            if (totalPoints >= 1000 && !achievements.includes('points1000')) {
                newAchievements.push(ACHIEVEMENTS.points1000);
            }
            if (totalPoints >= 5000 && !achievements.includes('points5000')) {
                newAchievements.push(ACHIEVEMENTS.points5000);
            }
            
            // Show new achievements
            newAchievements.forEach(achievement => {
                setTimeout(() => showAchievement(achievement), 500);
            });
            
            // Update display
            updatePointsDisplay();
            
            // Show points earned animation (calculate approximate points)
            // The actual calculation is done on the backend
            const pointsEarned = 10; // Base points - actual calculation is on backend
            showPointsEarned(pointsEarned);
        }, 500);
    }
    
    // Show points earned animation
    function showPointsEarned(points) {
        const animation = document.createElement('div');
        animation.className = 'points-earned';
        animation.textContent = `+${points}`;
        
        const pointsDisplay = document.querySelector('.points-value');
        if (pointsDisplay) {
            const rect = pointsDisplay.getBoundingClientRect();
            animation.style.left = rect.left + 'px';
            animation.style.top = rect.top + 'px';
        }
        
        document.body.appendChild(animation);
        
        setTimeout(() => {
            animation.classList.add('animate');
        }, 10);
        
        setTimeout(() => {
            animation.remove();
        }, 2000);
    }
    
    // Initialize gamification
    async function init() {
        // Fetch initial data
        await fetchGamificationData();
        
        updatePointsDisplay();
        
        // Hook into Done.completeTask to trigger gamification update
        const originalCompleteTask = window.exports.Done.completeTask;
        window.exports.Done.completeTask = function(taskUUID) {
            originalCompleteTask.call(this, taskUUID);
            
            // Trigger gamification update after task is completed
            setTimeout(() => onTaskComplete(), 100);
        };
    }
    
    // Export gamification API
    window.Gamification = {
        init: init,
        updatePointsDisplay: updatePointsDisplay,
        onTaskComplete: onTaskComplete
    };
    
})(window);