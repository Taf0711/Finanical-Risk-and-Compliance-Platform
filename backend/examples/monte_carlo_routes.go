// Example routes for Monte Carlo integration
// Add this to your main route setup file

package main

import (
	"github.com/Taf0711/financial-risk-monitor/internal/handlers"
	"github.com/Taf0711/financial-risk-monitor/internal/middleware"
	"github.com/Taf0711/financial-risk-monitor/internal/services"
	"github.com/gofiber/fiber/v2"
)

func setupMonteCarloRoutes(app *fiber.App, portfolioService *services.PortfolioService, authService *services.AuthService) {
	// Create Monte Carlo handler
	monteCarloHandler := handlers.NewMonteCarloHandler(portfolioService)

	// Monte Carlo API routes group
	mcApi := app.Group("/api/v1/monte-carlo")

	// Apply JWT middleware
	mcApi.Use(middleware.JWTMiddleware(authService))

	// Monte Carlo simulation endpoints
	mcApi.Post("/portfolios/:portfolio_id/simulate", monteCarloHandler.RunSimulation)
	mcApi.Post("/portfolios/:portfolio_id/validate", monteCarloHandler.RunQuickValidation)
	mcApi.Post("/portfolios/:portfolio_id/compare", monteCarloHandler.CompareSimulations)
	mcApi.Get("/portfolios/:portfolio_id/history", monteCarloHandler.GetSimulationHistory)
	mcApi.Get("/simulations/:simulation_id/status", monteCarloHandler.GetSimulationStatus)
}

/*
Usage Examples:

1. Run Full Monte Carlo Simulation:
POST /api/v1/monte-carlo/portfolios/{portfolio_id}/simulate
Content-Type: application/json
Authorization: Bearer {jwt_token}

{
  "num_simulations": 10000,
  "time_horizon_days": 22,
  "confidence_level": 0.95,
  "num_workers": 4,
  "market_regime": "NORMAL",
  "correlation_matrix": true
}

2. Quick Validation:
POST /api/v1/monte-carlo/portfolios/{portfolio_id}/validate
Authorization: Bearer {jwt_token}

3. Compare Scenarios:
POST /api/v1/monte-carlo/portfolios/{portfolio_id}/compare
Content-Type: application/json
Authorization: Bearer {jwt_token}

{
  "scenarios": [
    {
      "num_simulations": 5000,
      "time_horizon_days": 22,
      "market_regime": "NORMAL"
    },
    {
      "num_simulations": 5000,
      "time_horizon_days": 22,
      "market_regime": "STRESSED"
    }
  ]
}

4. Get Simulation History:
GET /api/v1/monte-carlo/portfolios/{portfolio_id}/history?limit=10
Authorization: Bearer {jwt_token}

Response Examples:

1. Simulation Result:
{
  "success": true,
  "portfolio": { ... },
  "simulation": {
    "config": { ... },
    "start_time": "2025-08-28T10:00:00Z",
    "end_time": "2025-08-28T10:00:02Z",
    "duration": "2.1s",
    "portfolio_value": 542500.00,
    "var_estimates": {
      "VaR_95": 11279.28,
      "VaR_99": 15896.42
    },
    "expected_shortfall": {
      "ES_95": 14413.20,
      "ES_99": 18769.40
    },
    "volatility": 0.2126,
    "max_drawdown": 0.1866,
    "skewness": 0.1163,
    "kurtosis": 0.3824,
    "liquidity_metrics": {
      "average_liquidity_score": 0.76,
      "time_to_liquidate": {
        "normal_market": 1.2,
        "stressed_market": 2.4,
        "crisis_market": 4.8
      }
    },
    "accuracy_validation": {
      "backtesting_results": {
        "95": 100.0,
        "99": 95.2
      },
      "var_violation_rate": {
        "95": 0.048,
        "99": 0.012
      }
    },
    "performance_metrics": {
      "simulation_time": "2.1s",
      "simulations_per_second": 4762,
      "memory_usage": 2670592
    }
  }
}

2. Quick Validation Result:
{
  "success": true,
  "validation": {
    "portfolio_value": 542500.00,
    "var_95": 11279.28,
    "var_99": 15896.42,
    "volatility": 21.26,
    "max_drawdown": 18.66,
    "liquidity_score": 0.76,
    "simulation_time": "65ms",
    "performance": {
      "simulations_per_sec": 15375,
      "memory_usage_mb": 2.54
    }
  }
}

Frontend Integration Example (JavaScript):

// Run Monte Carlo simulation
async function runMonteCarloSimulation(portfolioId, config = {}) {
  const defaultConfig = {
    num_simulations: 5000,
    time_horizon_days: 22,
    confidence_level: 0.95,
    num_workers: 4,
    market_regime: "NORMAL",
    correlation_matrix: true
  };

  try {
    const response = await fetch(`/api/v1/monte-carlo/portfolios/${portfolioId}/simulate`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${getJWTToken()}`
      },
      body: JSON.stringify({ ...defaultConfig, ...config })
    });

    const result = await response.json();

    if (result.success) {
      displaySimulationResults(result.simulation);
    } else {
      showError(result.error);
    }
  } catch (error) {
    showError('Failed to run simulation: ' + error.message);
  }
}

// Quick validation for dashboard
async function quickValidation(portfolioId) {
  try {
    const response = await fetch(`/api/v1/monte-carlo/portfolios/${portfolioId}/validate`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${getJWTToken()}`
      }
    });

    const result = await response.json();

    if (result.success) {
      updateDashboardMetrics(result.validation);
    }
  } catch (error) {
    console.error('Validation failed:', error);
  }
}

function displaySimulationResults(simulation) {
  // Update UI with simulation results
  document.getElementById('var-95').textContent =
    `$${simulation.var_estimates.VaR_95.toFixed(2)}`;
  document.getElementById('var-99').textContent =
    `$${simulation.var_estimates.VaR_99.toFixed(2)}`;
  document.getElementById('volatility').textContent =
    `${(simulation.volatility * 100).toFixed(2)}%`;
  document.getElementById('max-drawdown').textContent =
    `${(simulation.max_drawdown * 100).toFixed(2)}%`;

  if (simulation.liquidity_metrics) {
    document.getElementById('liquidity-score').textContent =
      simulation.liquidity_metrics.average_liquidity_score.toFixed(2);
  }

  // Show performance metrics
  document.getElementById('simulation-time').textContent = simulation.duration;
  document.getElementById('simulations-per-sec').textContent =
    simulation.performance_metrics.simulations_per_second.toFixed(0);
}

*/
