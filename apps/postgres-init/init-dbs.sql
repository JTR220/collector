-- Bases supplementaires pour price-tracker-service et notification-service.
-- Execute par postgres uniquement au premier demarrage (volume vierge) :
-- apres modification, faire `docker compose down -v` une fois.
CREATE DATABASE collector_price;
CREATE DATABASE collector_notifications;
