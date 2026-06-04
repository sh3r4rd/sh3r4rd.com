# ---------------------------------------------------------------------------
# API Gateway REST API
# ---------------------------------------------------------------------------

resource "aws_api_gateway_rest_api" "api" {
  name        = var.api_name
  description = "Recruiter Dashboard REST API"

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}

# ---------------------------------------------------------------------------
# Resources
# ---------------------------------------------------------------------------

resource "aws_api_gateway_resource" "recruiters" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_rest_api.api.root_resource_id
  path_part   = "recruiters"
}

resource "aws_api_gateway_resource" "recruiters_id" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.recruiters.id
  path_part   = "{id}"
}

resource "aws_api_gateway_resource" "stats" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_rest_api.api.root_resource_id
  path_part   = "stats"
}

# ---------------------------------------------------------------------------
# /recruiters — GET method + Lambda proxy integration
# ---------------------------------------------------------------------------

resource "aws_api_gateway_method" "recruiters_get" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  resource_id   = aws_api_gateway_resource.recruiters.id
  http_method   = "GET"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "recruiters_get" {
  rest_api_id             = aws_api_gateway_rest_api.api.id
  resource_id             = aws_api_gateway_resource.recruiters.id
  http_method             = aws_api_gateway_method.recruiters_get.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = var.api_handler_invoke_arn
}

# /recruiters — OPTIONS method (CORS preflight)

resource "aws_api_gateway_method" "recruiters_options" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  resource_id   = aws_api_gateway_resource.recruiters.id
  http_method   = "OPTIONS"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "recruiters_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.recruiters.id
  http_method = aws_api_gateway_method.recruiters_options.http_method
  type        = "MOCK"

  request_templates = {
    "application/json" = jsonencode({ statusCode = 200 })
  }
}

resource "aws_api_gateway_method_response" "recruiters_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.recruiters.id
  http_method = aws_api_gateway_method.recruiters_options.http_method
  status_code = "200"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = true
    "method.response.header.Access-Control-Allow-Methods" = true
    "method.response.header.Access-Control-Allow-Origin"  = true
  }

  response_models = {
    "application/json" = "Empty"
  }
}

resource "aws_api_gateway_integration_response" "recruiters_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.recruiters.id
  http_method = aws_api_gateway_method.recruiters_options.http_method
  status_code = aws_api_gateway_method_response.recruiters_options.status_code

  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = "'Content-Type'"
    "method.response.header.Access-Control-Allow-Methods" = "'GET,OPTIONS'"
    "method.response.header.Access-Control-Allow-Origin"  = "'${var.cors_allowed_origin}'"
  }
}

# ---------------------------------------------------------------------------
# /recruiters/{id} — GET method + Lambda proxy integration
# ---------------------------------------------------------------------------

resource "aws_api_gateway_method" "recruiters_id_get" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  resource_id   = aws_api_gateway_resource.recruiters_id.id
  http_method   = "GET"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "recruiters_id_get" {
  rest_api_id             = aws_api_gateway_rest_api.api.id
  resource_id             = aws_api_gateway_resource.recruiters_id.id
  http_method             = aws_api_gateway_method.recruiters_id_get.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = var.api_handler_invoke_arn
}

# /recruiters/{id} — OPTIONS method (CORS preflight)

resource "aws_api_gateway_method" "recruiters_id_options" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  resource_id   = aws_api_gateway_resource.recruiters_id.id
  http_method   = "OPTIONS"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "recruiters_id_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.recruiters_id.id
  http_method = aws_api_gateway_method.recruiters_id_options.http_method
  type        = "MOCK"

  request_templates = {
    "application/json" = jsonencode({ statusCode = 200 })
  }
}

resource "aws_api_gateway_method_response" "recruiters_id_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.recruiters_id.id
  http_method = aws_api_gateway_method.recruiters_id_options.http_method
  status_code = "200"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = true
    "method.response.header.Access-Control-Allow-Methods" = true
    "method.response.header.Access-Control-Allow-Origin"  = true
  }

  response_models = {
    "application/json" = "Empty"
  }
}

resource "aws_api_gateway_integration_response" "recruiters_id_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.recruiters_id.id
  http_method = aws_api_gateway_method.recruiters_id_options.http_method
  status_code = aws_api_gateway_method_response.recruiters_id_options.status_code

  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = "'Content-Type'"
    "method.response.header.Access-Control-Allow-Methods" = "'GET,OPTIONS'"
    "method.response.header.Access-Control-Allow-Origin"  = "'${var.cors_allowed_origin}'"
  }
}

# ---------------------------------------------------------------------------
# /stats — GET method + Lambda proxy integration
# ---------------------------------------------------------------------------

resource "aws_api_gateway_method" "stats_get" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  resource_id   = aws_api_gateway_resource.stats.id
  http_method   = "GET"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "stats_get" {
  rest_api_id             = aws_api_gateway_rest_api.api.id
  resource_id             = aws_api_gateway_resource.stats.id
  http_method             = aws_api_gateway_method.stats_get.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = var.api_handler_invoke_arn
}

# /stats — OPTIONS method (CORS preflight)

resource "aws_api_gateway_method" "stats_options" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  resource_id   = aws_api_gateway_resource.stats.id
  http_method   = "OPTIONS"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "stats_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.stats.id
  http_method = aws_api_gateway_method.stats_options.http_method
  type        = "MOCK"

  request_templates = {
    "application/json" = jsonencode({ statusCode = 200 })
  }
}

resource "aws_api_gateway_method_response" "stats_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.stats.id
  http_method = aws_api_gateway_method.stats_options.http_method
  status_code = "200"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = true
    "method.response.header.Access-Control-Allow-Methods" = true
    "method.response.header.Access-Control-Allow-Origin"  = true
  }

  response_models = {
    "application/json" = "Empty"
  }
}

resource "aws_api_gateway_integration_response" "stats_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.stats.id
  http_method = aws_api_gateway_method.stats_options.http_method
  status_code = aws_api_gateway_method_response.stats_options.status_code

  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = "'Content-Type'"
    "method.response.header.Access-Control-Allow-Methods" = "'GET,OPTIONS'"
    "method.response.header.Access-Control-Allow-Origin"  = "'${var.cors_allowed_origin}'"
  }
}

# ---------------------------------------------------------------------------
# Deployment + Stage
# ---------------------------------------------------------------------------

resource "aws_api_gateway_deployment" "api" {
  rest_api_id = aws_api_gateway_rest_api.api.id

  triggers = {
    redeployment = sha1(jsonencode([
      aws_api_gateway_resource.recruiters,
      aws_api_gateway_resource.recruiters_id,
      aws_api_gateway_resource.stats,
      aws_api_gateway_method.recruiters_get,
      aws_api_gateway_method.recruiters_options,
      aws_api_gateway_integration.recruiters_get,
      aws_api_gateway_integration.recruiters_options,
      aws_api_gateway_method_response.recruiters_options,
      aws_api_gateway_integration_response.recruiters_options,
      aws_api_gateway_method.recruiters_id_get,
      aws_api_gateway_method.recruiters_id_options,
      aws_api_gateway_integration.recruiters_id_get,
      aws_api_gateway_integration.recruiters_id_options,
      aws_api_gateway_method_response.recruiters_id_options,
      aws_api_gateway_integration_response.recruiters_id_options,
      aws_api_gateway_method.stats_get,
      aws_api_gateway_method.stats_options,
      aws_api_gateway_integration.stats_get,
      aws_api_gateway_integration.stats_options,
      aws_api_gateway_method_response.stats_options,
      aws_api_gateway_integration_response.stats_options,
    ]))
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_stage" "prod" {
  deployment_id = aws_api_gateway_deployment.api.id
  rest_api_id   = aws_api_gateway_rest_api.api.id
  stage_name    = "prod"
}

resource "aws_api_gateway_method_settings" "all" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  stage_name  = aws_api_gateway_stage.prod.stage_name
  method_path = "*/*"

  settings {
    throttling_burst_limit = var.throttling_burst_limit
    throttling_rate_limit  = var.throttling_rate_limit
  }
}

# ---------------------------------------------------------------------------
# Lambda permission — allow API Gateway to invoke the handler
# ---------------------------------------------------------------------------

resource "aws_lambda_permission" "api_gateway" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = var.api_handler_function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.api.execution_arn}/${aws_api_gateway_stage.prod.stage_name}/*"
}

# ---------------------------------------------------------------------------
# Custom domain — dedicated subdomain fronting the prod stage
# ---------------------------------------------------------------------------

# ACM certificate for the custom domain (REGIONAL — must be in the API's region).
resource "aws_acm_certificate" "api" {
  domain_name       = var.custom_domain_name
  validation_method = "DNS"

  lifecycle {
    create_before_destroy = true
  }
}

# DNS records that prove domain ownership to ACM.
resource "aws_route53_record" "cert_validation" {
  for_each = {
    for dvo in aws_acm_certificate.api.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  }

  zone_id         = var.hosted_zone_id
  name            = each.value.name
  type            = each.value.type
  records         = [each.value.record]
  ttl             = 60
  allow_overwrite = true
}

# Blocks until the certificate is validated and ISSUED.
resource "aws_acm_certificate_validation" "api" {
  certificate_arn         = aws_acm_certificate.api.arn
  validation_record_fqdns = [for record in aws_route53_record.cert_validation : record.fqdn]
}

# Regional custom domain, using the validated certificate.
resource "aws_api_gateway_domain_name" "api" {
  domain_name              = var.custom_domain_name
  regional_certificate_arn = aws_acm_certificate_validation.api.certificate_arn
  security_policy          = "TLS_1_2"

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}

# Map the domain root to the prod stage (no base path).
resource "aws_api_gateway_base_path_mapping" "api" {
  domain_name = aws_api_gateway_domain_name.api.domain_name
  api_id      = aws_api_gateway_rest_api.api.id
  stage_name  = aws_api_gateway_stage.prod.stage_name
}

# Alias the custom domain to the API Gateway regional endpoint.
resource "aws_route53_record" "api_alias" {
  zone_id = var.hosted_zone_id
  name    = var.custom_domain_name
  type    = "A"

  alias {
    name                   = aws_api_gateway_domain_name.api.regional_domain_name
    zone_id                = aws_api_gateway_domain_name.api.regional_zone_id
    evaluate_target_health = false
  }
}
