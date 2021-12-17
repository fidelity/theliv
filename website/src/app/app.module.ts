import { NgModule, CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { HttpClientModule, HTTP_INTERCEPTORS  } from '@angular/common/http';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { FontAwesomeModule } from '@fortawesome/angular-fontawesome';
import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { WindowToken, windowProvider } from './shared/util/window';
import { NavigationBarComponent } from './components/navigation-bar/navigation-bar.component';
import { KubernetesService } from './services/kubernetes.service';
import { KubePlatformComponent } from './kube-platform/kube-platform.component';
import { ResourceGroupContentComponent } from './components/resource-group-content/resource-group-content.component';
import { ReportCardComponent } from './components/report-card/report-card.component';
import { SocialUtil } from './shared/util/social-util';
import { KafkaComponent } from './kafka/kafka.component';
import { UnauthorizedInterceptor } from './shared/util/unauthorized-interceptor';
import { MatDialogModule } from '@angular/material/dialog';
import { ErrorDialog } from './shared/errors/error-dialog.component';
import { MatAutocompleteModule } from '@angular/material/autocomplete';
import { MatInputModule } from '@angular/material/input'
import { MatSelectModule } from '@angular/material/select'

@NgModule({
  declarations: [
    AppComponent,
    NavigationBarComponent,
    KubePlatformComponent,
    ResourceGroupContentComponent,
    ReportCardComponent,
    KafkaComponent,
    ErrorDialog
  ],
  imports: [
    BrowserModule,
    HttpClientModule,
    FormsModule,
    ReactiveFormsModule,
    AppRoutingModule,
    BrowserAnimationsModule,
    FontAwesomeModule,
    MatDialogModule,
    MatInputModule,
    MatAutocompleteModule,
    MatSelectModule,
  ],
  providers: [
    KubernetesService,
    SocialUtil,
    { provide: WindowToken, useFactory: windowProvider },
    {
      provide: HTTP_INTERCEPTORS,
      useClass: UnauthorizedInterceptor,
      multi: true
    }
  ],
  schemas: [CUSTOM_ELEMENTS_SCHEMA],
  bootstrap: [
    AppComponent
  ]
})
export class AppModule { }
