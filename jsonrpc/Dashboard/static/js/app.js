/**
 * INSPINIA - Responsive Admin Theme
 *
 */
(function () {
    angular.module('starter', [
        'ui.router',                    // Routing
        'oc.lazyLoad',                  // ocLazyLoad
        'ui.bootstrap',                 // Ui Bootstrap
        'pascalprecht.translate',       // Angular Translate
        'ngIdle',                       // Idle timer
        'ngSanitize',                   // ngSanitize
        'ngResource',                   // ngResource
        'ngCookies'                     // ngCookies
    ])
})();

// Other libraries are loaded dynamically in the config.js file using the library ocLazyLoad