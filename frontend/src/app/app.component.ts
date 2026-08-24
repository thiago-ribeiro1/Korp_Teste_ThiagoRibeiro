import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';
import { RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';
import { HealthService, StatusServicos } from './core/services/health.service';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, RouterLink, RouterLinkActive, CommonModule],
  templateUrl: './app.component.html',
  styleUrl: './app.component.scss',
})
export class AppComponent implements OnInit {
  status: StatusServicos = { estoque: true, faturamento: true };
  sidebarAberta = false;

  constructor(private health: HealthService) {}

  ngOnInit(): void {
    this.verificarStatus();
    // checa de novo periodicamente pra refletir quando o serviço volta do ar
    setInterval(() => this.verificarStatus(), 15000);
  }

  alternarSidebar(): void {
    this.sidebarAberta = !this.sidebarAberta;
  }

  fecharSidebar(): void {
    this.sidebarAberta = false;
  }

  private verificarStatus(): void {
    this.health.verificar().subscribe((status) => (this.status = status));
  }
}
