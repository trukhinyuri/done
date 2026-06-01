// Done gamification system - database-backed version
"use strict";

(function(window) {
    
    // Level thresholds
    const LEVELS = [
        { level: 1, points: 0, title: "Done" },
        { level: 2, points: 100, title: "Apprentice" },
        { level: 3, points: 200, title: "Journeyman" },
        { level: 4, points: 300, title: "Expert" },
        { level: 5, points: 400, title: "Master" },
        { level: 6, points: 500, title: "Champion" },
        { level: 7, points: 600, title: "Hero" },
        { level: 8, points: 700, title: "Legend" },
        { level: 9, points: 800, title: "Mythic" },
        { level: 10, points: 900, title: "Deity" }
    ];
    
    // Achievements
    const ACHIEVEMENTS = {
        firstTask: { id: 'firstTask', name: 'First Steps', description: 'Complete your first task', icon: 'First' },
        streak3: { id: 'streak3', name: 'On Fire', description: '3 day streak', icon: '3d' },
        streak7: { id: 'streak7', name: 'Week Warrior', description: '7 day streak', icon: '7d' },
        streak30: { id: 'streak30', name: 'Monthly Master', description: '30 day streak', icon: '30d' },
        points1000: { id: 'points1000', name: 'Point Collector', description: 'Earn 1000 points', icon: '1k' },
        points5000: { id: 'points5000', name: 'Point Master', description: 'Earn 5000 points', icon: '5k' },
        speedDemon: { id: 'speedDemon', name: 'Speed Demon', description: 'Complete 5 tasks in one day', icon: '5/day' },
        earlyBird: { id: 'earlyBird', name: 'Early Bird', description: 'Complete a task before deadline', icon: 'Early' }
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
                    <span class="tasks-icon">Completed</span>
                    <span class="tasks-value">${completedTasks}</span>
                </div>
                <div class="points-score">
                    <span class="points-icon">Points</span>
                    <span class="points-value">${totalPoints.toLocaleString()}</span>
                </div>
                <div class="level-progress">
                    <div class="progress-bar">
                        <div class="progress-fill" style="width: ${progress.percent}%"></div>
                    </div>
                    <div class="progress-text">${progress.pointsNeeded} points to next level${progress.daysToNext !== null && progress.daysToNext > 0 ? ` (~${progress.daysToNext} day${progress.daysToNext !== 1 ? 's' : ''})` : ''}</div>
                </div>
                <button class="sound-toggle ${soundEnabled ? 'enabled' : ''}" title="Toggle sound effects">
                    <span class="sound-icon">${soundEnabled ? 'Sound on' : 'Sound off'}</span>
                </button>
            </div>
        `;
        
        // Add sound toggle event listener
        const soundToggle = pointsDisplay.querySelector('.sound-toggle');
        if (soundToggle) {
            soundToggle.addEventListener('click', function() {
                const enabled = window.NinstyleSounds.toggleSound();
                this.classList.toggle('enabled', enabled);
                this.querySelector('.sound-icon').textContent = enabled ? 'Sound on' : 'Sound off';
            });
        }
    }
    
    // Handle task completion - Points are now calculated on the backend
    async function onTaskComplete() {
        const previousPoints = cachedData ? (cachedData.total_points || 0) : 0;
        const previousAchievements = cachedData && Array.isArray(cachedData.achievements) ? cachedData.achievements : [];

        // Wait a bit for the backend to update
        setTimeout(async () => {
            // Clear cache to force refresh
            cachedData = null;
            
            // Fetch updated data
            const data = await fetchGamificationData();
            
            const totalPoints = data.total_points || 0;
            const achievements = Array.isArray(data.achievements) ? data.achievements : [];
            const newAchievements = achievements
                .filter(achievementID => !previousAchievements.includes(achievementID))
                .map(achievementID => ACHIEVEMENTS[achievementID])
                .filter(Boolean);
            
            // Show new achievements
            newAchievements.forEach(achievement => {
                setTimeout(() => showAchievement(achievement), 500);
            });
            
            // Update display
            updatePointsDisplay();
            
            const pointsEarned = totalPoints - previousPoints;
            if (pointsEarned > 0) {
                showPointsEarned(pointsEarned);
            }
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
