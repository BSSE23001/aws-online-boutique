#!/usr/bin/python
from concurrent import futures
import os
import time
import grpc
from jinja2 import Environment, FileSystemLoader, select_autoescape, TemplateError
from google.api_core.exceptions import GoogleAPICallError

import demo_pb2
import demo_pb2_grpc
from grpc_health.v1 import health_pb2
from grpc_health.v1 import health_pb2_grpc



import smtplib
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart




from logger import getJSONLogger
logger = getJSONLogger('emailservice-server')

# Loads confirmation email template from file
env = Environment(
    loader=FileSystemLoader('templates'),
    autoescape=select_autoescape(['html', 'xml'])
)
template = env.get_template('confirmation.html')

class BaseEmailService(demo_pb2_grpc.EmailServiceServicer):
  def Check(self, request, context):
    return health_pb2.HealthCheckResponse(
      status=health_pb2.HealthCheckResponse.SERVING)
  
  def Watch(self, request, context):
    return health_pb2.HealthCheckResponse(
      status=health_pb2.HealthCheckResponse.UNIMPLEMENTED)

class EmailService(BaseEmailService):
  def __init__(self):
    # raise Exception('cloud mail client not implemented')
    super().__init__()

  @staticmethod
  # def send_email(client, email_address, content):
  def send_email(email_address, content):
    # --- CONFIGURATION ---
    sender_email = "magdalena.daugherty28@ethereal.email"
    sender_password = "vHRpCA8AP33EDJvFuk"
    smtp_server = "smtp.ethereal.email"
    smtp_port = 587
    # ---------------------

    msg = MIMEMultipart()
    msg['From'] = sender_email
    msg['To'] = email_address
    msg['Subject'] = "Order Confirmation (Ethereal Test)"

    # Attach the rendered HTML content from Jinja2
    msg.attach(MIMEText(content, 'html'))

    try:
        server = smtplib.SMTP(smtp_server, smtp_port)
        server.starttls() # Secure the connection
        server.login(sender_email, sender_password)
        server.sendmail(sender_email, email_address, msg.as_string())
        server.quit()
        logger.info("Actual email sent to: {}".format(email_address))
    except Exception as e:
        logger.error("Failed to send email: {}".format(e))
    # response = client.send_message(
    #   sender = client.sender_path(project_id, region, sender_id),
    #   envelope_from_authority = '',
    #   header_from_authority = '',
    #   envelope_from_address = from_address,
    #   simple_message = {
    #     "from": {
    #       "address_spec": from_address,
    #     },
    #     "to": [{
    #       "address_spec": email_address
    #     }],
    #     "subject": "Your Confirmation Email",
    #     "html_body": content
    #   }
    # )
    # logger.info("Message sent: {}".format(response.rfc822_message_id))

  def SendOrderConfirmation(self, request, context):
    email = request.email
    order = request.order

    try:
      confirmation = template.render(order = order)
    except TemplateError as err:
      context.set_details("An error occurred when preparing the confirmation mail.")
      logger.error(err.message)
      context.set_code(grpc.StatusCode.INTERNAL)
      return demo_pb2.Empty()

    try:
      # EmailService.send_email(self.client, email, confirmation)
      EmailService.send_email(email, confirmation)
    except GoogleAPICallError as err:
      context.set_details("An error occurred when sending the email.")
      print(err.message)
      context.set_code(grpc.StatusCode.INTERNAL)
      return demo_pb2.Empty()

    return demo_pb2.Empty()

# class DummyEmailService(BaseEmailService):
#   def SendOrderConfirmation(self, request, context):
#     logger.info('A request to send order confirmation email to {} has been received.'.format(request.email))
#     return demo_pb2.Empty()

class HealthCheck():
  def Check(self, request, context):
    return health_pb2.HealthCheckResponse(
      status=health_pb2.HealthCheckResponse.SERVING)

# def start(dummy_mode):
def start():
  server = grpc.server(futures.ThreadPoolExecutor(max_workers=10),)
  service = EmailService()
  # service = None
  # if dummy_mode:
  #   service = DummyEmailService()
  # else:
  #   raise Exception('non-dummy mode not implemented yet')

  demo_pb2_grpc.add_EmailServiceServicer_to_server(service, server)
  health_pb2_grpc.add_HealthServicer_to_server(service, server)

  port = os.environ.get('PORT', "8080")
  logger.info("listening on port: "+port)
  server.add_insecure_port('[::]:'+port)
  server.start()
  try:
    while True:
      time.sleep(3600)
  except KeyboardInterrupt:
    server.stop(0)

if __name__ == '__main__':
  logger.info('starting the email service in dummy mode.')
  
  # start(dummy_mode = True)
  start()
