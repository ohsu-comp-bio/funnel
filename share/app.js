app = angular.module('TESApp', ['ngRoute'])

app.controller('JobListController', function($scope, $http) {
		"use strict";

		$scope.url = "/v1/jobs";
		$scope.tasks = [];

		$scope.fetchContent = function() {
			$http.get($scope.url).then(function(result){
				$scope.jobs = result.data.jobs;
			});
		}

		$scope.fetchContent();
});

app.controller('WorkerListController', function($scope, $http) {
		"use strict";

		$scope.url = "/v1/jobs-service";
		$scope.workers = [];

		$scope.fetchContent = function() {
			$http.get($scope.url).then(function(result){
				$scope.workers = result.data;
			});
		}

		$scope.fetchContent();
});

app.controller('JobInfoController',
    function($scope, $http, $routeParams) {
        $scope.url = "/v1/jobs/" + $routeParams.job_id

        $scope.job_info = {};
        $scope.fetchContent = function() {
            $http.get($scope.url).then(function(result){
                $scope.job_info = result.data
            })
        }
        $scope.fetchContent();
    }
);

app.config(['$routeProvider',
    function($routeProvider) {
        $routeProvider.when('/', {
           templateUrl: 'static/list.html',
        }).
        when('/jobs/:job_id', {
           templateUrl: 'static/jobs.html'
        })
    }
]);